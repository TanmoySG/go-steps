package funcs

import "fmt"

var itr int = 0

func Add(args ...any) ([]interface{}, error) {
	fmt.Printf("Adding %v\n", args)
	return []interface{}{args[0].(int) + args[1].(int)}, nil
}

func Sub(args ...any) ([]interface{}, error) {
	fmt.Printf("Sub %v\n", args)
	return []interface{}{args[0].(int) - args[1].(int)}, nil
}

func Multiply(args ...any) ([]interface{}, error) {
	fmt.Printf("Multiply %v\n", args)
	return []interface{}{args[0].(int) * args[1].(int)}, nil
}

func Divide(args ...any) ([]interface{}, error) {
	fmt.Printf("Divide %v\n", args)
	return []interface{}{args[0].(int) / args[1].(int)}, nil
}

// Step will error 3times and return arg*30 and arg*31 on the 4th try
func StepWillError3Times(args ...any) ([]interface{}, error) {
	fmt.Printf("Running fake error function for arg [%v]\n", args)
	if itr == 3 {
		return []interface{}{args[0].(int) * 30, args[0].(int) * 50}, nil
	}

	itr += 1
	return nil, fmt.Errorf("error to retry")
}
