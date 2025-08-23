// 현재는 standard library를 구현하지 않았기 때문에 c runtime에서 선언하여 사용한다.
#include <stdio.h>

void print(int value) {
    printf("🚀 Naviary says: %d\n", value);
}

void printBool(int value) {
    printf("🚀 Naviary says: %s\n", value ? "true" : "false");
}

void printString(char* value) {
    printf("🚀 Naviary says: %s\n", value);
}