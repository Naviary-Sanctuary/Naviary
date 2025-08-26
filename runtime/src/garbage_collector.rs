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
    // 일단 연결리스트로 구현한다.
    first_object: *mut ObjectHeader,

    total_bytes_allocated: usize,
    garbage_collection_threshold: usize,
}

impl GarbageCollector {
    pub fn new() -> Self {
        GarbageCollector {
            // NULL 포인터를 만듬.
            first_object: ptr::null_mut(),
            total_bytes_allocated: 0,
            garbage_collection_threshold: 1024 * 1024, // 1MB
        }
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

#[cfg(test)]
mod test_with_header {
    use super::*;

    #[test]
    fn test_linked_list_creation() {
        let mut gc = GarbageCollector::new();

        // 첫 번째 객체 할당
        let data1 = gc.allocate(100);
        assert!(!data1.is_null());

        // 두 번째 객체 할당
        let data2 = gc.allocate(200);
        assert!(!data2.is_null());

        // 세 번째 객체 할당
        let data3 = gc.allocate(50);
        assert!(!data3.is_null());

        // 연결 리스트 확인
        unsafe {
            // 데이터 포인터에서 헤더 포인터 구하기
            let header3 = (data3 as *mut ObjectHeader).sub(1);
            let header2 = (data2 as *mut ObjectHeader).sub(1);
            let header1 = (data1 as *mut ObjectHeader).sub(1);

            // 가장 최근 객체가 first_object여야 함
            assert_eq!(gc.first_object, header3);

            // header3 -> header2 -> header1 -> null 순서
            assert_eq!((*header3).next_object, header2);
            assert_eq!((*header2).next_object, header1);
            assert!((*header1).next_object.is_null());
        }
    }

    #[test]
    fn test_object_size_tracking() {
        let mut gc = GarbageCollector::new();

        let _data1 = gc.allocate(100);
        let expected = ObjectHeader::HEADER_SIZE + 100;
        assert_eq!(gc.total_bytes_allocated, expected);

        let _data2 = gc.allocate(200);
        let expected = expected + ObjectHeader::HEADER_SIZE + 200;
        assert_eq!(gc.total_bytes_allocated, expected);
    }
}

#[cfg(test)]
mod test_sweep {
    use super::*;

    #[test]
    fn test_sweep_unmarked_objects() {
        let mut gc = GarbageCollector::new();

        // 3개 객체 할당
        let data1 = gc.allocate(100);
        let _data2 = gc.allocate(200);
        let data3 = gc.allocate(300);

        // 수동으로 일부만 mark
        unsafe {
            let header1 = (data1 as *mut ObjectHeader).sub(1);
            let header3 = (data3 as *mut ObjectHeader).sub(1);

            (*header1).is_marked = true; // 1번 살림
            // 2번은 mark 안 함 (죽임)
            (*header3).is_marked = true; // 3번 살림
        }

        let before = gc.total_bytes_allocated;

        // Sweep 실행!
        gc.sweep();

        let after = gc.total_bytes_allocated;

        // 2번 객체 크기만큼 줄어야 함
        let expected_freed = ObjectHeader::HEADER_SIZE + 200;
        assert_eq!(before - after, expected_freed);

        // 연결 리스트 확인 (3 -> 1 -> null)
        unsafe {
            let header1 = (data1 as *mut ObjectHeader).sub(1);
            let header3 = (data3 as *mut ObjectHeader).sub(1);

            assert_eq!(gc.first_object, header3);
            assert_eq!((*header3).next_object, header1);
            assert!((*header1).next_object.is_null());
        }
    }

    #[test]
    fn test_sweep_all_unmarked() {
        let mut gc = GarbageCollector::new();

        // 3개 할당, 아무것도 mark 안 함
        let _data1 = gc.allocate(100);
        let _data2 = gc.allocate(200);
        let _data3 = gc.allocate(300);

        // 모두 해제되어야 함
        gc.sweep();

        assert_eq!(gc.total_bytes_allocated, 0);
        assert!(gc.first_object.is_null());
    }
}
