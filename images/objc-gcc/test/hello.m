#import <Foundation/Foundation.h>

int main(int argc, char *argv[])
{
    NSString *str = @"Hello World";
    printf("%s", [str cString]);
    return 0;
}
