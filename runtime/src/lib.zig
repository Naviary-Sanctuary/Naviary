const io = @import("io.zig");

// TODO: add types
pub export fn print_int(value: i64) void {
    io.naviary_print_int(value);
}

pub export fn print_string(string_pointer: [*:0]const u8) void {
    io.naviary_print_string(string_pointer);
}
