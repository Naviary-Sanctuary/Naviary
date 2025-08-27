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
    garbage_collection_threshold: usize,
    root_objects: Vec<*mut ObjectHeader>,
}

impl GarbageCollector {
    pub fn new() -> Self {
        GarbageCollector {
            // NULL 포인터를 만듬.
            first_object: ptr::null_mut(),
            total_bytes_allocated: 0,
            garbage_collection_threshold: 1024 * 1024, // 1MB
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

        if self.should_collect(total_size) {
            self.collect();

            self.garbage_collection_threshold = std::cmp::max(self.total_bytes_allocated * 2, 1024);
        }

        //메모리 할당
        // 저수준 메모리 할당 API -> C의 malloc/free와 비슷함
        let layout = Layout::from_size_align(total_size, ObjectHeader::HEADER_ALIGN)
            .expect("Failed to create layout");

        let header_ptr = unsafe {
            alloc(layout) as *mut ObjectHeader // 메모리 할당
        };

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

    fn should_collect(&self, size: usize) -> bool {
        self.total_bytes_allocated + size >= self.garbage_collection_threshold
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
fn test_automatic_gc_trigger() {
    let mut gc = GarbageCollector::new();

    // 임계값을 낮게 설정 (테스트용)
    gc.garbage_collection_threshold = 1000; // 1KB

    println!("초기 임계값: {} bytes", gc.garbage_collection_threshold);

    // 루트 없이 계속 할당 (모두 가비지)
    for i in 0..10 {
        let size = 200; // 각 200 bytes
        let _obj = gc.allocate(size);

        println!("할당 #{}: 총 {} bytes", i, gc.total_bytes_allocated);

        // 5번째쯤에 임계값 초과 → 자동 GC 발생!
    }

    // GC가 실행되어서 메모리가 정리되었을 것
    assert!(gc.total_bytes_allocated < 1000);
}

#[test]
fn test_gc_with_mixed_roots() {
    let mut gc = GarbageCollector::new();
    gc.garbage_collection_threshold = 2000; // 2KB

    let mut roots = Vec::new();

    // 20개 객체 할당 (일부만 루트로 유지)
    for i in 0..20 {
        let obj = gc.allocate(150);

        // 짝수 번째만 루트로 유지
        if i % 2 == 0 {
            gc.add_root(obj);
            roots.push(obj);
        }
        // 홀수 번째는 가비지가 됨
    }

    println!("GC 후 남은 객체: {}", roots.len());

    // 루트 정리
    for obj in roots {
        gc.remove_root(obj);
    }
}
