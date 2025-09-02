use std::{
    alloc::{Layout, alloc, realloc},
    mem,
};

#[repr(C)]
#[derive(Debug, Clone, Copy, PartialEq)]
pub enum ObjectType {
    String,
    IntArray,
    FloatArray,
    BoolArray,
    StringArray,
    // TODO: AnyArray
}

// - 필드 순서를 우리가 정한대로 보장함
// - 포인터 연산으로 헤더와 데이터 사이를 이동 가능
// - 메모리 정렬을 보장함
#[repr(C)]
pub struct ObjectHeader {
    pub is_marked: bool,

    // 가변포인터를 사용하는 이유
    // - 마지막 객체는 null
    // - 나중에 이 필드가 수정되어야함
    pub next_object: *mut ObjectHeader,

    // 헤더 + 데이터 사이즈
    pub object_size: usize,

    pub object_type: ObjectType,
}

impl ObjectHeader {
    pub const HEADER_SIZE: usize = mem::size_of::<ObjectHeader>();
    // 헤더 정렬 요구사항
    pub const HEADER_ALIGN: usize = mem::align_of::<ObjectHeader>();
}

#[cfg(target_pointer_width = "32")]
pub type NaviaryInt = i32;

#[cfg(target_pointer_width = "64")]
pub type NaviaryInt = i64;

pub type NaviaryFloat = f64;

#[repr(C)]
pub struct StringObject {
    pub header: ObjectHeader,
    pub length: usize, // 문자열 길이
}

impl StringObject {
    pub unsafe fn get_chars(&self) -> &[u8] {
        unsafe {
            let data_ptr = (self as *const _ as *const u8).add(mem::size_of::<StringObject>());
            std::slice::from_raw_parts(data_ptr, self.length)
        }
    }
    pub unsafe fn to_str(&self) -> &str {
        unsafe { std::str::from_utf8_unchecked(self.get_chars()) }
    }
}

#[repr(C)]
pub struct IntArrayObject {
    pub header: ObjectHeader,
    pub length: usize,
    pub capacity: usize,
    pub elements: *mut NaviaryInt,
}

impl IntArrayObject {
    pub unsafe fn get(&self, index: usize) -> NaviaryInt {
        if index >= self.length {
            panic!("Array index out of bounds {} >= {}", index, self.length);
        }

        unsafe { *self.elements.add(index) }
    }

    pub unsafe fn set(&self, index: usize, value: NaviaryInt) {
        if index >= self.length {
            panic!("Array index out of bounds {} >= {}", index, self.length);
        }

        unsafe {
            *self.elements.add(index) = value;
        }
    }

    pub unsafe fn push(&mut self, value: NaviaryInt) {
        if self.length + 1 >= self.capacity {
            unsafe {
                self.grow();
            }
        }

        unsafe {
            *self.elements.add(self.length) = value;
        }
        self.length += 1;
    }

    pub unsafe fn pop(&mut self) -> Option<NaviaryInt> {
        if self.length == 0 {
            return None;
        }
        unsafe {
            self.length -= 1;
            Some(*self.elements.add(self.length))
        }
    }

    unsafe fn grow(&mut self) {
        let new_capacity = match self.capacity {
            0 => 4,
            _ if self.capacity < 1024 => self.capacity * 2,
            _ => self.capacity + (self.capacity / 2),
        };

        unsafe {
            self.resize(new_capacity);
        }
    }

    pub unsafe fn resize(&mut self, new_capacity: usize) {
        if new_capacity < self.capacity {
            panic!("Array capacity cannot be decreased");
        }

        if new_capacity == self.capacity {
            return;
        }

        let new_layout =
            Layout::array::<NaviaryInt>(new_capacity).expect("Failed to create layout");

        let new_elements = if self.elements.is_null() || self.capacity == 0 {
            unsafe { alloc(new_layout) as *mut NaviaryInt }
        } else {
            let old_layout =
                Layout::array::<NaviaryInt>(self.capacity).expect("Failed to create layout");
            unsafe {
                realloc(self.elements as *mut u8, old_layout, new_layout.size()) as *mut NaviaryInt
            }
        };

        if new_elements.is_null() {
            panic!("Array allocation failed: Out of Memory");
        }

        self.elements = new_elements;
        self.capacity = new_capacity;
    }

    pub fn len(&self) -> usize {
        self.length
    }
}

#[repr(C)]
pub struct FloatArrayObject {
    pub header: ObjectHeader,
    pub length: usize,
    pub capacity: usize,
    pub elements: *mut NaviaryFloat,
}

impl FloatArrayObject {
    // 요소 접근 헬퍼
    pub unsafe fn get(&self, index: usize) -> NaviaryFloat {
        if index >= self.length {
            panic!("Array index out of bounds: {} >= {}", index, self.length);
        }
        unsafe { *self.elements.add(index) }
    }

    pub unsafe fn set(&mut self, index: usize, value: NaviaryFloat) {
        if index >= self.length {
            panic!("Array index out of bounds: {} >= {}", index, self.length);
        }
        unsafe {
            *self.elements.add(index) = value;
        }
    }

    pub unsafe fn push(&mut self, value: NaviaryFloat) {
        if self.length + 1 >= self.capacity {
            unsafe {
                self.grow();
            }
        }
        unsafe {
            *self.elements.add(self.length) = value;
        }
        self.length += 1;
    }

    pub unsafe fn pop(&mut self) -> Option<NaviaryFloat> {
        if self.length == 0 {
            return None;
        }
        unsafe {
            self.length -= 1;
            Some(*self.elements.add(self.length))
        }
    }
    unsafe fn grow(&mut self) {
        let new_capacity = match self.capacity {
            0 => 4,
            _ if self.capacity < 1024 => self.capacity * 2,
            _ => self.capacity + (self.capacity / 2),
        };

        unsafe {
            self.resize(new_capacity);
        }
    }

    pub unsafe fn resize(&mut self, new_capacity: usize) {
        if new_capacity < self.capacity {
            panic!("Array capacity cannot be decreased");
        }

        if new_capacity == self.capacity {
            return;
        }

        let new_layout =
            Layout::array::<NaviaryFloat>(new_capacity).expect("Failed to create layout");

        let new_elements = if self.elements.is_null() || self.capacity == 0 {
            unsafe { alloc(new_layout) as *mut NaviaryFloat }
        } else {
            let old_layout =
                Layout::array::<NaviaryFloat>(self.capacity).expect("Failed to create layout");
            unsafe {
                realloc(self.elements as *mut u8, old_layout, new_layout.size())
                    as *mut NaviaryFloat
            }
        };

        if new_elements.is_null() {
            panic!("Array allocation failed: Out of Memory");
        }

        self.elements = new_elements;
        self.capacity = new_capacity;
    }

    pub fn len(&self) -> usize {
        self.length
    }
}

#[repr(C)]
pub struct BoolArrayObject {
    pub header: ObjectHeader,
    pub length: usize,
    pub capacity: usize,
    pub elements: *mut bool,
}

impl BoolArrayObject {
    // 요소 접근 헬퍼
    pub unsafe fn get(&self, index: usize) -> bool {
        if index >= self.length {
            panic!("Array index out of bounds: {} >= {}", index, self.length);
        }
        unsafe { *self.elements.add(index) }
    }

    pub unsafe fn set(&mut self, index: usize, value: bool) {
        if index >= self.length {
            panic!("Array index out of bounds: {} >= {}", index, self.length);
        }
        unsafe {
            *self.elements.add(index) = value;
        }
    }

    pub unsafe fn push(&mut self, value: bool) {
        if self.length + 1 >= self.capacity {
            unsafe {
                self.grow();
            }
        }
        unsafe {
            *self.elements.add(self.length) = value;
        }
        self.length += 1;
    }

    pub unsafe fn pop(&mut self) -> Option<bool> {
        if self.length == 0 {
            return None;
        }
        unsafe {
            self.length -= 1;
            Some(*self.elements.add(self.length))
        }
    }
    unsafe fn grow(&mut self) {
        let new_capacity = match self.capacity {
            0 => 4,
            _ if self.capacity < 1024 => self.capacity * 2,
            _ => self.capacity + (self.capacity / 2),
        };

        unsafe {
            self.resize(new_capacity);
        }
    }

    pub unsafe fn resize(&mut self, new_capacity: usize) {
        if new_capacity < self.capacity {
            panic!("Array capacity cannot be decreased");
        }

        if new_capacity == self.capacity {
            return;
        }

        let new_layout = Layout::array::<bool>(new_capacity).expect("Failed to create layout");

        let new_elements = if self.elements.is_null() || self.capacity == 0 {
            unsafe { alloc(new_layout) as *mut bool }
        } else {
            let old_layout = Layout::array::<bool>(self.capacity).expect("Failed to create layout");
            unsafe { realloc(self.elements as *mut u8, old_layout, new_layout.size()) as *mut bool }
        };

        if new_elements.is_null() {
            panic!("Array allocation failed: Out of Memory");
        }

        self.elements = new_elements;
        self.capacity = new_capacity;
    }

    pub fn len(&self) -> usize {
        self.length
    }
}

#[repr(C)]
pub struct StringArrayObject {
    pub header: ObjectHeader,
    pub length: usize,
    pub capacity: usize,
    pub elements: *mut *mut StringObject,
}

impl StringArrayObject {
    pub unsafe fn get(&self, index: usize) -> *mut StringObject {
        if index >= self.length {
            panic!("Array index out of bounds: {} >= {}", index, self.length);
        }

        unsafe { *self.elements.add(index) }
    }

    pub unsafe fn set(&mut self, index: usize, value: *mut StringObject) {
        if index >= self.length {
            panic!("Array index out of bounds: {} >= {}", index, self.length);
        }

        unsafe { *self.elements.add(index) = value };
    }

    pub unsafe fn push(&mut self, value: *mut StringObject) {
        if self.length + 1 >= self.capacity {
            unsafe {
                self.grow();
            }
        }
        unsafe {
            *self.elements.add(self.length) = value;
        }
        self.length += 1;
    }

    pub unsafe fn pop(&mut self) -> Option<*mut StringObject> {
        if self.length == 0 {
            None
        } else {
            self.length -= 1;
            unsafe { Some(*self.elements.add(self.length)) }
        }
    }

    unsafe fn grow(&mut self) {
        let new_capacity = if self.capacity == 0 {
            4
        } else if self.capacity < 1024 {
            self.capacity * 2
        } else {
            self.capacity + self.capacity / 2
        };

        unsafe { self.resize(new_capacity) };
    }

    pub unsafe fn resize(&mut self, new_capacity: usize) {
        if new_capacity < self.length {
            panic!("Cannot resize below current length");
        }

        if new_capacity == self.capacity {
            return;
        }

        let new_layout =
            Layout::array::<*mut StringObject>(new_capacity).expect("Layout calculation failed");

        let new_elements = if self.elements.is_null() || self.capacity == 0 {
            unsafe { alloc(new_layout) as *mut *mut StringObject }
        } else {
            let old_layout = Layout::array::<*mut StringObject>(self.capacity)
                .expect("Layout calculation failed");
            unsafe {
                realloc(self.elements as *mut u8, old_layout, new_layout.size())
                    as *mut *mut StringObject
            }
        };

        if new_elements.is_null() {
            panic!("Failed to resize array: Out of Memory");
        }

        // null로 초기화
        if new_capacity > self.capacity {
            unsafe {
                std::ptr::write_bytes(
                    new_elements.add(self.capacity),
                    0,
                    new_capacity - self.capacity,
                )
            };
        }

        self.elements = new_elements;
        self.capacity = new_capacity;
    }

    pub fn len(&self) -> usize {
        self.length
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_dynamic_growth() {
        // GC 없이 직접 테스트 (실제로는 GC 통해서만 할당해야 함)
        unsafe {
            let mut array = IntArrayObject {
                header: ObjectHeader {
                    is_marked: false,
                    next_object: std::ptr::null_mut(),
                    object_size: std::mem::size_of::<IntArrayObject>(),
                    object_type: ObjectType::IntArray,
                },
                length: 0,
                capacity: 0, // 작은 초기 용량
                elements: alloc(Layout::array::<NaviaryInt>(2).unwrap()) as *mut NaviaryInt,
            };

            array.push(1);
            array.push(2);
            array.push(3);
            array.push(4);
            array.push(5); // 여기서 자동 확장!
            assert_eq!(array.capacity, 8);
            assert_eq!(array.len(), 5);

            array.push(6);
            array.push(7);
            array.push(8);
            array.push(9); // 또 확장!
            assert_eq!(array.capacity, 16);

            // 값 확인
            assert_eq!(array.get(0), 1);
            assert_eq!(array.get(1), 2);
            assert_eq!(array.get(2), 3);
            assert_eq!(array.get(3), 4);
            assert_eq!(array.get(4), 5);
            assert_eq!(array.get(5), 6);
            assert_eq!(array.get(6), 7);
            assert_eq!(array.get(7), 8);
            assert_eq!(array.get(8), 9);
            assert_eq!(array.len(), 9);

            // pop 테스트
            assert_eq!(array.pop(), Some(9));
            assert_eq!(array.len(), 8); // pop을 했기 때문에 length가 줄어듦

            // 메모리 해제 (실제로는 GC가 처리)
            if !array.elements.is_null() {
                let layout = Layout::array::<NaviaryInt>(array.capacity).unwrap();
                std::alloc::dealloc(array.elements as *mut u8, layout);
            }
        }
    }
}
