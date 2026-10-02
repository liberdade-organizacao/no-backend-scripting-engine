use wasm_bindgem::prelude::*;

#[wasm_bindgen]
pub fn boot(ptr: u32, size: u32) -> u32 {
    return ptr + size;
}

