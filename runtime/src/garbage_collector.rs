use super::object::{ObjectHeader, ObjectType, StringObject};
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

    pub fn allocate_string(&mut self, text: &str) -> *mut StringObject {
        let object_size = mem::size_of::<StringObject>();
        let size = object_size + text.len();

        if self.should_collect(size) {
            self.collect();
        }

        let layout = Layout::from_size_align(size, 8).unwrap();
        let ptr = unsafe { alloc(layout) as *mut StringObject };

        if ptr.is_null() {
            panic!("String allocation failed: Out of Memory");
        }

        unsafe {
            (*ptr).header = ObjectHeader {
                is_marked: false,
                next_object: self.first_object,
                object_size: size,
                object_type: ObjectType::String,
            };
            (*ptr).length = text.len();

            let data_ptr = (ptr as *mut u8).add(object_size);

            std::ptr::copy_nonoverlapping(text.as_ptr(), data_ptr, text.len());

            self.first_object = &mut (*ptr).header as *mut ObjectHeader;
        }

        self.total_bytes_allocated += size;

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

#[cfg(test)]
mod test_object_allocation {
    use super::*;
    #[test]
    fn test_allocate_string() {
        let mut gc = GarbageCollector::new();

        // Naviary: let name = "Hello";
        let str_obj = gc.allocate_string("Hello");

        assert!(!str_obj.is_null());

        unsafe {
            // 길이 확인
            assert_eq!((*str_obj).length, 5);

            // 타입 확인
            assert_eq!((*str_obj).header.object_type, ObjectType::String);

            // 문자 데이터 읽기
            let chars = (*str_obj).get_chars();
            let text = std::str::from_utf8(chars).unwrap();
            assert_eq!(text, "Hello");

            println!("String object info:");
            println!("  Address: {:p}", str_obj);
            println!("  Value: \"{}\"", text);
            println!("  Length: {}", (*str_obj).length);
            println!("  Total size: {} bytes", (*str_obj).header.object_size);
        }
    }

    #[test]
    fn test_multiple_strings() {
        let mut gc = GarbageCollector::new();

        let str1 = gc.allocate_string("Hello");
        let str2 = gc.allocate_string("World");
        let str3 = gc.allocate_string("Naviary");

        // 다른 주소인지 확인
        assert_ne!(str1, str2);
        assert_ne!(str2, str3);

        unsafe {
            // 각각 올바른 값 저장
            assert_eq!((*str1).to_str(), "Hello");
            assert_eq!((*str2).to_str(), "World");
            assert_eq!((*str3).to_str(), "Naviary");

            // 크기 확인
            let str1_size = std::mem::size_of::<StringObject>() + 5; // "Hello"
            let str2_size = std::mem::size_of::<StringObject>() + 5; // "World"
            let str3_size = std::mem::size_of::<StringObject>() + 7; // "Naviary"

            assert_eq!((*str1).header.object_size, str1_size);
            assert_eq!((*str2).header.object_size, str2_size);
            assert_eq!((*str3).header.object_size, str3_size);
        }
    }

    #[test]
    fn test_empty_string() {
        let mut gc = GarbageCollector::new();

        // 빈 문자열
        let empty = gc.allocate_string("");

        unsafe {
            assert_eq!((*empty).length, 0);
            assert_eq!((*empty).to_str(), "");

            // 크기는 StringObject 구조체 크기만
            assert_eq!(
                (*empty).header.object_size,
                std::mem::size_of::<StringObject>()
            );
        }
    }

    #[test]
    fn test_string_with_gc() {
        let mut gc = GarbageCollector::new();
        gc.garbage_collection_threshold = 200; // 낮게 설정

        // 여러 문자열 할당
        let str1 = gc.allocate_string("Keep me"); // 7 chars
        let _str2 = gc.allocate_string("Delete me"); // 9 chars
        let str3 = gc.allocate_string("Keep too"); // 8 chars
        let _str4 = gc.allocate_string("Delete too"); // 10 chars

        // 일부만 루트로 등록
        gc.add_root(str1 as *mut u8);
        gc.add_root(str3 as *mut u8);

        println!("Before GC: {} bytes", gc.total_bytes_allocated);

        // GC 실행
        gc.collect();

        println!("After GC: {} bytes", gc.total_bytes_allocated);

        // 살아남은 문자열 확인
        unsafe {
            assert_eq!((*str1).to_str(), "Keep me");
            assert_eq!((*str3).to_str(), "Keep too");
            // str2, str4는 접근하면 안 됨 (해제됨)
        }

        // 메모리 계산
        let expected = std::mem::size_of::<StringObject>() + 7 +  // "Keep me"
        std::mem::size_of::<StringObject>() + 8; // "Keep too"

        assert_eq!(gc.total_bytes_allocated, expected);

        // 정리
        gc.remove_root(str1 as *mut u8);
        gc.remove_root(str3 as *mut u8);
    }
}
