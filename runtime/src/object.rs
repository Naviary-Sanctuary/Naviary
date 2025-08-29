use std::mem;

#[repr(C)]
#[derive(Debug, Clone, Copy, PartialEq)]
pub enum ObjectType {
    Integer,
    Float,
    String,
    Boolean,
    Array,
    Nil,
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

#[repr(C)]
pub struct IntegerObject {
    pub header: ObjectHeader,
    pub value: NaviaryInt,
}

#[repr(C)]
pub struct FloatObject {
    pub header: ObjectHeader,
    pub value: f64,
}

#[repr(C)]
pub struct BooleanObject {
    pub header: ObjectHeader,
    pub value: bool,
}

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
pub struct ArrayObject {
    pub header: ObjectHeader,
    pub length: usize,
    pub capacity: usize,
    pub elements: *mut *mut ObjectHeader,
}

impl ArrayObject {
    pub unsafe fn get_elements(&self, index: usize) -> *mut ObjectHeader {
        if index >= self.length {
            panic!("Array index out of bounds");
        }

        unsafe { *self.elements.add(index) }
    }

    pub unsafe fn set_element(&self, index: usize, value: *mut ObjectHeader) {
        if index >= self.length {
            panic!("Array index out of bounds");
        }

        unsafe {
            *self.elements.add(index) = value;
        }
    }
}

#[repr(C)]
pub struct NilObject {
    pub header: ObjectHeader,
}
