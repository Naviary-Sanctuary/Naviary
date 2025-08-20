// 현재는 standard library를 구현하지 않았기 때문에 c runtime에서 선언하여 사용한다.
#include <stdio.h>

void print(int value) {
    printf("🚀 Naviary says: %d\n", value);
    
    // 디버깅 정보 추가
    fprintf(stderr, "[DEBUG] printed value: %d\n", value);
}