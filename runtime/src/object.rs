use std::mem;

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

    pub unsafe fn push(&self, value: NaviaryInt) {
        if self.length >= self.capacity {
            panic!("Array capacity reached");
            //TODO: resize
        }

        unsafe {
            *self.elements.add(self.length) = value;
        }
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

    // 요소 추가 (나중에 구현)
    pub unsafe fn push(&mut self, value: NaviaryFloat) {
        if self.length >= self.capacity {
            panic!("Array is full, resize needed");
            // TODO: resize
        }
        unsafe {
            *self.elements.add(self.length) = value;
        }
        self.length += 1;
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

    // 요소 추가 (나중에 구현)
    pub unsafe fn push(&mut self, value: bool) {
        if self.length >= self.capacity {
            panic!("Array is full, resize needed");
            // TODO: resize
        }
        unsafe {
            *self.elements.add(self.length) = value;
        }
        self.length += 1;
    }
}

#[repr(C)]
pub struct StringArrayObject {
    pub header: ObjectHeader,
    pub length: usize,
    pub capacity: usize,
    pub elements: *mut StringObject,
}

impl StringArrayObject {
    pub unsafe fn get(&self, index: usize) -> *mut StringObject {
        if index >= self.length {
            panic!("Array index out of bounds: {} >= {}", index, self.length);
        }

        unsafe { self.elements.add(index) }
    }

    pub unsafe fn set(&mut self, index: usize, value: StringObject) {
        if index >= self.length {
            panic!("Array index out of bounds: {} >= {}", index, self.length);
        }

        unsafe {
            *self.elements.add(index) = value;
        }
    }

    // 요소 추가 (나중에 구현)
    pub unsafe fn push(&mut self, value: StringObject) {
        if self.length >= self.capacity {
            panic!("Array is full, resize needed");
            // TODO: resize
        }
        unsafe {
            *self.elements.add(self.length) = value;
        }
        self.length += 1;
    }
}
