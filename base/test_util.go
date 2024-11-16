package base

// ================================================================================ //
//                                                                                  //
//  测试工具库                                                                        //
//  @author anonymous                                                               //
//  @updated_at 2024.11.15 01:04:49                                                 //
//                                                                                  //
//  @cmd_help:                                                                      //
//  1. unit test:                                                                   //
//     $ go test xxx                                                                //
//  2. bench test:                                                                  //
//     $ go test -benchmem -run=^$ -bench ^<$function_name>$ -count=<$count> -v     //
//                                                                                  //
//                                                                                  //
// ================================================================================ //

// TestFn 测试用例函数
type TestFn func() string
