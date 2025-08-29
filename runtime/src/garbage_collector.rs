use super::object::{IntegerObject, NaviaryInt, ObjectHeader, ObjectType};
use std::{
    alloc::{Layout, alloc, dealloc},
    mem, ptr,
};

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

        let header_ptr = data_ptr as *mut ObjectHeader;

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
        self.garbage_collection_threshold = std::cmp::max(self.total_bytes_allocated * 2, 1024);
    }

    pub fn allocate_integer(&mut self, value: NaviaryInt) -> *mut IntegerObject {
        let size = mem::size_of::<IntegerObject>();

        if self.should_collect(size) {
            self.collect();
        }

        let layout = Layout::from_size_align(size, 8).unwrap();
        let ptr = unsafe { alloc(layout) as *mut IntegerObject };

        if ptr.is_null() {
            panic!("Integer 할당 실패: Out of Memory");
        }

        unsafe {
            (*ptr) = IntegerObject {
                header: ObjectHeader {
                    is_marked: false,
                    next_object: self.first_object,
                    object_size: size,
                    object_type: ObjectType::Integer,
                },
                value,
            };

            self.first_object = &mut (*ptr).header as *mut ObjectHeader;
        }

        self.total_bytes_allocated += size;

        println!("Integer({}) 할당: {:p}, {} bytes", value, ptr, size);

        ptr
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

// src/runtime/garbage_collector.rs의 테스트 모듈에 추가

#[cfg(test)]
mod test_object_allocation {
    use super::*;

    #[test]
    fn test_allocate_integer() {
        let mut gc = GarbageCollector::new();

        // Naviary: let x = 42;
        let int_obj = gc.allocate_integer(42);

        // null이 아닌지 확인
        assert!(!int_obj.is_null());

        unsafe {
            // 값이 제대로 저장되었는지
            assert_eq!((*int_obj).value, 42);

            // 타입이 맞는지
            assert_eq!((*int_obj).header.object_type, ObjectType::Integer);

            // 크기가 맞는지
            assert_eq!(
                (*int_obj).header.object_size,
                std::mem::size_of::<IntegerObject>()
            );

            println!("Integer 객체 정보:");
            println!("  주소: {:p}", int_obj);
            println!("  값: {}", (*int_obj).value);
            println!("  타입: {:?}", (*int_obj).header.object_type);
            println!("  크기: {} bytes", (*int_obj).header.object_size);
        }

        // 메모리 사용량 확인
        assert_eq!(
            gc.total_bytes_allocated,
            std::mem::size_of::<IntegerObject>()
        );
    }

    #[test]
    fn test_multiple_integers() {
        let mut gc = GarbageCollector::new();

        // 여러 개 할당
        let int1 = gc.allocate_integer(10);
        let int2 = gc.allocate_integer(20);
        let int3 = gc.allocate_integer(30);

        // 다른 주소인지 확인
        assert_ne!(int1, int2);
        assert_ne!(int2, int3);

        unsafe {
            // 값들이 제대로 저장되었는지
            assert_eq!((*int1).value, 10);
            assert_eq!((*int2).value, 20);
            assert_eq!((*int3).value, 30);

            // 연결 리스트 확인
            // 가장 최근 할당된 int3가 first_object여야 함
            assert_eq!(gc.first_object, &(*int3).header as *const _ as *mut _);

            // int3 -> int2 -> int1 -> null 순서
            assert_eq!(
                (*int3).header.next_object,
                &(*int2).header as *const _ as *mut _
            );
            assert_eq!(
                (*int2).header.next_object,
                &(*int1).header as *const _ as *mut _
            );
            assert!((*int1).header.next_object.is_null());
        }

        // 총 메모리 확인
        assert_eq!(
            gc.total_bytes_allocated,
            std::mem::size_of::<IntegerObject>() * 3
        );
    }

    #[test]
    fn test_integer_with_gc() {
        let mut gc = GarbageCollector::new();

        // 임계값을 충분히 크게 (GC 자동 실행 방지)
        gc.garbage_collection_threshold = 1000;

        let int_size = std::mem::size_of::<IntegerObject>();
        println!("IntegerObject 크기: {} bytes", int_size);

        // 10개 할당
        let mut roots = Vec::new();

        for i in 0..10 {
            let int_obj = gc.allocate_integer(i * 10);

            // 짝수만 루트로 등록 (0, 20, 40, 60, 80)
            if i % 2 == 0 {
                // IntegerObject를 u8 포인터로 변환할 때 주의!
                // header의 주소가 아니라 객체의 주소를 전달
                gc.add_root(int_obj as *mut u8);
                roots.push(int_obj);
                println!("루트 등록: Integer({})", i * 10);
            }
        }

        println!("할당된 객체 수: 10");
        println!("루트 객체 수: {}", roots.len());
        println!("GC 전 메모리: {} bytes", gc.total_bytes_allocated);

        // 수동 GC 실행
        gc.collect();

        println!("GC 후 메모리: {} bytes", gc.total_bytes_allocated);

        // 살아남은 객체 확인
        unsafe {
            for (idx, &obj) in roots.iter().enumerate() {
                let expected_value = (idx * 2) * 10; // 0, 20, 40, 60, 80
                let actual_value = (*obj).value;

                println!(
                    "객체 {}: 예상={}, 실제={}",
                    idx, expected_value, actual_value
                );

                assert_eq!(
                    actual_value, expected_value as NaviaryInt,
                    "객체 {}의 값이 잘못됨",
                    idx
                );
            }
        }

        // 5개만 살아남아야 함
        let expected_objects = 5;
        let expected_memory = int_size * expected_objects;
        assert_eq!(
            gc.total_bytes_allocated, expected_memory,
            "메모리 크기가 예상과 다름"
        );
    }
}
