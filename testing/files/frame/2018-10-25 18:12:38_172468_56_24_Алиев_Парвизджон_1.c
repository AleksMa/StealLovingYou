#include <stdio.h>
#include <stdlib.h>
#include <string.h>

int main(int argc, char **argv) {

    if (argc != 4) {
        printf("Usage: frame <height> <width> <text>\n");
        return 0;
    }
    int x = atoi(argv[1]);
    int y = atoi(argv[2]);
    int lns = strlen(argv[3]);
    int i, j;
    int middle = x / 2;
    if(2*middle==x){
        middle-=1;
    }
    if ((x > 2) && (y > lns + 1)) {
        for (j = 0; j < y; j++){
            printf("*");
        }
        printf("\n");
        for (i = 1; i < x - 1; i++){
            if (i == middle) {
                printf("*");
                for (j = 1; j < (y-lns)/2; j++){
                    printf(" ");
                }
                printf("%s", argv[3]);
                for (j = 1; j < (y-lns+1)/2; j++){
                    printf(" ");
                }
                printf("*\n");
            }
            else {
                printf("*");
                for (j = 1; j <= y -2; j++){
                    printf(" ");
                }
                printf("*\n");
            }
        }

        for (j = 0; j < y; j++){
            if (j == (y - 1)){
                printf("*\n");
            }
            else printf("*");
        }
    }
    else printf("Error\n");
    return 0;
}
