package test

import (
	"fmt"
	"testing"
)

// ================================================================================ //
//                                                                                  //
//  测试模版示例                                                                      //
//  @author anonymous                                                               //
//  @updated_at 2024.11.16 20:46:08                                                 //
//                                                                                  //
//  @cmd_help:                                                                      //
//  1. unit test:                                                                   //
//     $ go test xxx                                                                //
//  2. bench test:                                                                  //
//     $ go test -benchmem -run=^$ -bench ^<$function_name>$ -count=<$count> -v     //
//                                                                                  //
//                                                                                  //
// ================================================================================ //

// TestExample 单元测试样例
func TestExample(t *testing.T) {
	fmt.Println("This is a test example")
}
