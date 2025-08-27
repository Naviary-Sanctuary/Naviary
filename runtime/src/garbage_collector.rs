use std::{
    alloc::{Layout, alloc, dealloc},
    mem, ptr,
};

// - 필드 순서를 우리가 정한대로 보장함
// - 포인터 연산으로 헤더와 데이터 사이를 이동 가능
// - 메모리 정렬을 보장함
#[repr(C)]
pub struct ObjectHeader {
    is_marked: bool,

    // 가변포인터를 사용하는 이유
    // - 마지막 객체는 null
    // - 나중에 이 필드가 수정되어야함
    next_object: *mut ObjectHeader,

    // 헤더 + 데이터 사이즈
    object_size: usize,
}

impl ObjectHeader {
    const HEADER_SIZE: usize = mem::size_of::<ObjectHeader>();

    // 헤더 정렬 요구사항
    const HEADER_ALIGN: usize = mem::align_of::<ObjectHeader>();
}

pub struct GarbageCollector {
    first_object: *mut ObjectHeader,
    total_bytes_allocated: usize,
    _garbage_collection_threshold: usize,
    root_objects: Vec<*mut ObjectHeader>,
}

impl GarbageCollector {
    pub fn new() -> Self {
        GarbageCollector {
            // NULL 포인터를 만듬.
            first_object: ptr::null_mut(),
            total_bytes_allocated: 0,
            _garbage_collection_threshold: 1024 * 1024, // 1MB
            root_objects: Vec::new(),
        }
    }

    pub fn add_root(&mut self, data_ptr: *mut u8) {
        if data_ptr.is_null() {
            return;
        }

        let header_ptr = unsafe { (data_ptr as *mut ObjectHeader).sub(1) };

        if !self.root_objects.contains(&header_ptr) {
            self.root_objects.push(header_ptr);
        }
    }

    pub fn remove_root(&mut self, data_ptr: *mut u8) {
        if data_ptr.is_null() {
            return;
        }

        let header_ptr = unsafe { (data_ptr as *mut ObjectHeader).sub(1) };
        self.root_objects.retain(|&root| root != header_ptr);
    }

    pub fn mark(&mut self) {
        // clone()하는 이유: 빌림 규칙
        for &root in &self.root_objects.clone() {
            self.mark_object(root);
        }
    }

    fn mark_object(&mut self, object: *mut ObjectHeader) {
        if object.is_null() {
            return;
        }

        unsafe {
            if (*object).is_marked {
                return;
            }

            (*object).is_marked = true;

            // TODO: 추후 참조 추적 구현
        }
    }

    pub fn collect(&mut self) {
        self.mark();
        self.sweep();
    }

    // 현재는 시스템 malloc 호출만 한다.
    pub fn allocate(&mut self, size: usize) -> *mut u8 {
        // 사이즈 계산
        let header_size = ObjectHeader::HEADER_SIZE;
        let total_size = size + header_size;

        //메모리 할당
        // 저수준 메모리 할당 API -> C의 malloc/free와 비슷함
        let layout = Layout::from_size_align(total_size, ObjectHeader::HEADER_ALIGN)
            .expect("Failed to create layout");

        let header_ptr = unsafe {
            alloc(layout) as *mut ObjectHeader // 메모리 할당
        };

        if header_ptr.is_null() {
            panic!("Memory allocation failed");
        }

        // 헤더 초기화
        unsafe {
            // *header_ptr는 뭔가요?
            // 포인터가 가리키는 메모리에 직접 쓰기
            // C의 *ptr = value와 같음
            (*header_ptr) = ObjectHeader {
                is_marked: false,               // 새 객체는 mark 안 됨
                next_object: self.first_object, // 기존 리스트 앞에 추가
                object_size: total_size,
            };
        }

        // linked 리스트 업데이트
        self.first_object = header_ptr;
        self.total_bytes_allocated += total_size;

        // 데이터 포인터 반환
        // 사용자는 헤더 다음부터 사용
        unsafe {
            let data_ptr = (header_ptr as *mut u8).add(header_size);
            data_ptr
        }
    }

    pub fn sweep(&mut self) {
        //이전 객체 추적 (linked list 수정용)
        let mut previous: *mut ObjectHeader = ptr::null_mut();
        let mut current_object = self.first_object;

        let mut freed_bytes = 0;

        unsafe {
            while !current_object.is_null() {
                let next = (*current_object).next_object;

                if (*current_object).is_marked {
                    (*current_object).is_marked = false;
                    previous = current_object;
                    current_object = next;
                } else {
                    let size = (*current_object).object_size;
                    freed_bytes += size;

                    if previous.is_null() {
                        self.first_object = next;
                    } else {
                        (*previous).next_object = next;
                    }

                    let layout = Layout::from_size_align(size, ObjectHeader::HEADER_ALIGN)
                        .expect("Failed to create layout");
                    dealloc(current_object as *mut u8, layout);

                    current_object = next;
                }
            }
        }

        self.total_bytes_allocated -= freed_bytes;
    }
}

#[test]
fn test_mark_and_sweep() {
    let mut gc = GarbageCollector::new();

    // 5개 객체 할당
    let obj1 = gc.allocate(100);
    let _obj2 = gc.allocate(200);
    let obj3 = gc.allocate(300);
    let _obj4 = gc.allocate(400);
    let obj5 = gc.allocate(500);

    // obj1, obj3, obj5만 루트로 등록
    // (obj2, obj4는 가비지가 될 예정)
    gc.add_root(obj1);
    gc.add_root(obj3);
    gc.add_root(obj5);

    // GC 실행!
    gc.collect();

    // 검증
    let expected_remaining = ObjectHeader::HEADER_SIZE + 100 +  // obj1
        ObjectHeader::HEADER_SIZE + 300 +  // obj3  
        ObjectHeader::HEADER_SIZE + 500; // obj5

    assert_eq!(gc.total_bytes_allocated, expected_remaining);

    // 살아남은 객체들 확인
    unsafe {
        let header1 = (obj1 as *mut ObjectHeader).sub(1);
        let header3 = (obj3 as *mut ObjectHeader).sub(1);
        let header5 = (obj5 as *mut ObjectHeader).sub(1);

        // mark는 해제되어야 함 (다음 GC 위해)
        assert!(!(*header1).is_marked);
        assert!(!(*header3).is_marked);
        assert!(!(*header5).is_marked);
    }
}
