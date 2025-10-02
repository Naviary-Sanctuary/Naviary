const io = @import("io.zig");

// TODO: add types
pub export fn print(value: c_int) void {
    io.naviary_print(value);
}
