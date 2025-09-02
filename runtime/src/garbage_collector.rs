use crate::object::{
    BoolArrayObject, FloatArrayObject, IntArrayObject, NaviaryFloat, NaviaryInt, StringArrayObject,
};

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

            match (*object).object_type {
                ObjectType::StringArray => {
                    let array = object as *mut StringArrayObject;
                    for i in 0..(*array).length {
                        let element = *(*array).elements.add(i);
                        if !element.is_null() {
                            self.mark_object(element as *mut ObjectHeader);
                        }
                    }
                }
                _ => {} // Primitive 배열은 추가 마킹 불필요
            }
        }
    }

    pub fn collect(&mut self) {
        self.mark();
        self.sweep();
        self.garbage_collection_threshold = std::cmp::max(self.total_bytes_allocated * 2, 1024);
    }

    fn should_collect(&self, size: usize) -> bool {
        self.total_bytes_allocated + size >= self.garbage_collection_threshold
    }

    pub fn sweep(&mut self) {
        let mut previous: *mut ObjectHeader = ptr::null_mut();
        let mut current = self.first_object;

        unsafe {
            while !current.is_null() {
                let next = (*current).next_object;

                if (*current).is_marked {
                    (*current).is_marked = false;
                    previous = current;
                    current = next;
                } else {
                    // 객체와 관련 메모리 해제
                    self.free_object(current);

                    if previous.is_null() {
                        self.first_object = next;
                    } else {
                        (*previous).next_object = next;
                    }

                    current = next;
                }
            }
        }
    }

    unsafe fn register_object(&mut self, object: *mut ObjectHeader) {
        unsafe {
            (*object).next_object = self.first_object;
            self.first_object = object;
        }
    }

    unsafe fn free_object(&mut self, object: *mut ObjectHeader) {
        unsafe {
            let object_size = (*object).object_size;

            // array의 경우 elements 버퍼를 먼저 해제해야 dangling pointer가 발생하지 않음
            match (*object).object_type {
                ObjectType::IntArray => {
                    let array = object as *mut IntArrayObject;
                    self.free_array_elements(
                        (*array).elements as *mut u8,
                        (*array).capacity,
                        mem::size_of::<NaviaryInt>(),
                    );
                }
                ObjectType::FloatArray => {
                    let array = object as *mut FloatArrayObject;
                    self.free_array_elements(
                        (*array).elements as *mut u8,
                        (*array).capacity,
                        mem::size_of::<NaviaryFloat>(),
                    );
                }
                ObjectType::BoolArray => {
                    let array = object as *mut BoolArrayObject;
                    self.free_array_elements(
                        (*array).elements as *mut u8,
                        (*array).capacity,
                        mem::size_of::<bool>(),
                    );
                }
                ObjectType::StringArray => {
                    let array = object as *mut StringArrayObject;
                    self.free_array_elements(
                        (*array).elements as *mut u8,
                        (*array).capacity,
                        mem::size_of::<*mut StringObject>(),
                    );
                }
                _ => {}
            }

            let layout = Layout::from_size_align(object_size, mem::align_of::<ObjectHeader>())
                .expect("Invalid layout");
            dealloc(object as *mut u8, layout);

            self.total_bytes_allocated -= object_size;
        }
    }

    unsafe fn free_array_elements(
        &mut self,
        elements: *mut u8,
        capacity: usize,
        element_size: usize,
    ) {
        if !elements.is_null() && capacity > 0 {
            let size = capacity * element_size;
            let layout =
                Layout::from_size_align(size, mem::align_of::<usize>()).expect("Invalid layout");
            unsafe {
                dealloc(elements, layout);
            }
            self.total_bytes_allocated -= size;
        }
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

            self.register_object(&mut (*ptr).header);
        }

        self.total_bytes_allocated += size;

        ptr
    }

    // ===== Int 배열 할당 =====
    pub fn allocate_int_array(&mut self, capacity: usize) -> *mut IntArrayObject {
        self.allocate_array::<IntArrayObject, NaviaryInt>(
            capacity,
            ObjectType::IntArray,
            |array, elements| unsafe {
                (*array).header.object_type = ObjectType::IntArray;
                (*array).length = 0;
                (*array).capacity = capacity;
                (*array).elements = elements;
            },
        )
    }

    // ===== Float 배열 할당 =====
    pub fn allocate_float_array(&mut self, capacity: usize) -> *mut FloatArrayObject {
        self.allocate_array::<FloatArrayObject, NaviaryFloat>(
            capacity,
            ObjectType::FloatArray,
            |array, elements| unsafe {
                (*array).header.object_type = ObjectType::FloatArray;
                (*array).length = 0;
                (*array).capacity = capacity;
                (*array).elements = elements;
            },
        )
    }

    // ===== Bool 배열 할당 =====
    pub fn allocate_bool_array(&mut self, capacity: usize) -> *mut BoolArrayObject {
        self.allocate_array::<BoolArrayObject, bool>(
            capacity,
            ObjectType::BoolArray,
            |array, elements| unsafe {
                (*array).header.object_type = ObjectType::BoolArray;
                (*array).length = 0;
                (*array).capacity = capacity;
                (*array).elements = elements;
            },
        )
    }

    // ===== String 배열 할당 =====
    pub fn allocate_string_array(&mut self, capacity: usize) -> *mut StringArrayObject {
        self.allocate_array::<StringArrayObject, *mut StringObject>(
            capacity,
            ObjectType::StringArray,
            |array, elements| unsafe {
                (*array).header.object_type = ObjectType::StringArray;
                (*array).length = 0;
                (*array).capacity = capacity;
                (*array).elements = elements;
            },
        )
    }

    pub fn allocate_array<ArrayType, ElementType>(
        &mut self,
        capacity: usize,
        object_type: ObjectType,
        init_fn: impl FnOnce(*mut ArrayType, *mut ElementType),
    ) -> *mut ArrayType {
        let array_size = mem::size_of::<ArrayType>();
        let element_size = mem::size_of::<ElementType>();
        let elements_size = capacity * element_size;
        let total_size = array_size + elements_size;

        // GC 체크
        if self.should_collect(total_size) {
            self.collect();
        }

        // 배열 객체 할당
        let array_layout = Layout::from_size_align(array_size, mem::align_of::<ArrayType>())
            .expect("Invalid array layout");
        let array_ptr = unsafe { alloc(array_layout) as *mut ArrayType };

        if array_ptr.is_null() {
            panic!("Array allocation failed: Out of Memory");
        }

        let elements_ptr = if capacity > 0 {
            let elements_layout =
                Layout::from_size_align(elements_size, mem::align_of::<ElementType>())
                    .expect("Invalid elements layout");

            let ptr = unsafe { alloc(elements_layout) as *mut ElementType };

            if ptr.is_null() {
                unsafe {
                    dealloc(array_ptr as *mut u8, array_layout);
                }
                panic!("Array elements allocation failed: Out of Memory");
            }

            unsafe {
                ptr::write_bytes(ptr, 0, capacity);
            }

            ptr
        } else {
            ptr::null_mut()
        };

        unsafe {
            let header_ptr = array_ptr as *mut ObjectHeader;
            (*header_ptr) = ObjectHeader {
                is_marked: false,
                next_object: ptr::null_mut(),
                object_size: array_size,
                object_type,
            };

            init_fn(array_ptr, elements_ptr);

            self.register_object(header_ptr);
        }

        self.total_bytes_allocated += total_size;
        array_ptr
    }
}
#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_allocate_int_array() {
        let mut gc = GarbageCollector::new();

        let array = gc.allocate_int_array(10);

        unsafe {
            assert!(!array.is_null());
            assert_eq!((*array).capacity, 10);
            assert_eq!((*array).length, 0);
            assert!(!(*array).elements.is_null());

            // 값 설정 테스트
            (*array).length = 3;
            (*array).set(0, 100);
            (*array).set(1, 200);
            (*array).set(2, 300);

            assert_eq!((*array).get(0), 100);
            assert_eq!((*array).get(1), 200);
            assert_eq!((*array).get(2), 300);
        }
    }

    #[test]
    fn test_allocate_empty_array() {
        let mut gc = GarbageCollector::new();

        let array = gc.allocate_int_array(0);

        unsafe {
            assert!(!array.is_null());
            assert_eq!((*array).capacity, 0);
            assert_eq!((*array).length, 0);
            assert!((*array).elements.is_null());
        }
    }

    #[test]
    fn test_array_gc() {
        let mut gc = GarbageCollector::new();
        gc.garbage_collection_threshold = 500;

        let int_array = gc.allocate_int_array(10);
        let _float_array = gc.allocate_float_array(10);
        let string_array = gc.allocate_string_array(5);

        // String 객체 생성하고 배열에 추가
        let str1 = gc.allocate_string("Hello");
        unsafe {
            (*string_array).length = 1;
            (*string_array).set(0, str1);
        }

        // 일부만 루트로 등록 (직접 캐스팅)
        gc.add_root(int_array as *mut u8);
        gc.add_root(string_array as *mut u8);

        let before = gc.total_bytes_allocated;
        gc.collect();
        let after = gc.total_bytes_allocated;

        assert!(after < before); // float_array가 해제됨

        unsafe {
            // 살아있는 객체 확인
            assert_eq!((*int_array).capacity, 10);
            assert_eq!((*string_array).capacity, 5);
            assert_eq!((*(*string_array).get(0)).to_str(), "Hello");
        }

        // 루트 제거
        gc.remove_root(int_array as *mut u8);
        gc.remove_root(string_array as *mut u8);
    }

    #[test]
    fn test_different_array_types() {
        let mut gc = GarbageCollector::new();

        let int_arr = gc.allocate_int_array(5);
        let float_arr = gc.allocate_float_array(5);
        let bool_arr = gc.allocate_bool_array(5);
        let string_arr = gc.allocate_string_array(5);

        unsafe {
            // 각 타입별로 값 설정
            (*int_arr).length = 1;
            (*int_arr).set(0, 42);

            (*float_arr).length = 1;
            (*float_arr).set(0, 3.14);

            (*bool_arr).length = 1;
            (*bool_arr).set(0, true);

            let str = gc.allocate_string("Test");
            (*string_arr).length = 1;
            (*string_arr).set(0, str);

            // 확인
            assert_eq!((*int_arr).get(0), 42);
            assert_eq!((*float_arr).get(0), 3.14);
            assert_eq!((*bool_arr).get(0), true);
            assert_eq!((*(*string_arr).get(0)).to_str(), "Test");
        }
    }
}
