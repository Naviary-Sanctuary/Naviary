const std = @import("std");

// TODO: int type only for now
export fn print(value: i32) void {
    std.debug.print("{}\n", .{value});
}
