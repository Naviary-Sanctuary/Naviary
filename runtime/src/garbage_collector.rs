pub struct GarbageCollector {
    total_bytes_allocated: usize,
}

impl GarbageCollector {
    pub fn new() -> Self {
        GarbageCollector {
            total_bytes_allocated: 0,
        }
    }

    // 현재는 시스템 malloc 호출만 한다.
    pub fn allocate(&mut self, size: usize) -> *mut u8 {
        println!("Allocating memory: {}", size);

        // 저수준 메모리 할당 API -> C의 malloc/free와 비슷함
        use std::alloc::{Layout, alloc};

        // 플랫폼별 기본 정렬 값 - 32bit: 4, 64bit: 8
        const ALIGNMENT: usize = std::mem::align_of::<usize>();
        // 메모리 할당 시 크기와 정렬을 지정하는 구조체
        // 정렬은 2의 제곱수여야함. (bytes단위 정렬이기 때문)
        // 실패하면 panic
        let layout = Layout::from_size_align(size, ALIGNMENT).unwrap();

        // alloc은 메모리 안정성을 보장하지 못하기 때문에 unsafe keyword를 사용해야한다.
        // - 초기화되지 않은 메모리 반환
        // - 해제 책임은 우리에게
        // - 잘못하면 크래시
        let pointer = unsafe {
            alloc(layout) // 메모리 할당
        };

        if pointer.is_null() {
            panic!("Memory allocation failed");
        }

        self.total_bytes_allocated += size;
        println!(
            "Successfully allocated. address: {:p}, total allocated memory: {} bytes",
            pointer, self.total_bytes_allocated
        );

        pointer
    }
}

#[cfg(test)]
mod test_garbage_collector {
    use super::GarbageCollector;

    #[test]
    fn test_basic_allocation() {
        let mut gc = GarbageCollector::new();

        let pointer1 = gc.allocate(100);
        assert!(!pointer1.is_null());

        let pointer2 = gc.allocate(200);
        assert!(!pointer2.is_null());

        // 두 포인터가 다른지 확인 (다른 메모리 영역)
        assert_ne!(pointer1, pointer2);
    }
}
