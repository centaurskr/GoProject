///
/// @brief  : GOLANG structure test
/// @file   : structure.go
/// @date   : 2024. 04. 30. (화) 16:21:16 KST
/// @author : Cento
///
package main
import "fmt"

type ClassInfo struct{
	Class int
	No    int
}

type Student struct {
	ClassInfo
	Name      string
}

type DupStudent struct{
	ClassInfo
	Name      string
	No        int
}


func main(){
	var s Student = Student{
		ClassInfo:ClassInfo{Class:1, No:1},
		Name : "John",
	}
	fmt.Println(`
	type ClassInfo struct{
		Class int
		No    int
	}
	var s Student = Student{
		ClassInfo:ClassInfo{Class:1, No:1},
		Name : "John",
	}
	`)
	fmt.Println("Class : ", s.Class)
	fmt.Println("No    : ", s.No)
	fmt.Println("Name  : ", s.Name)

	fmt.Println()

	fmt.Println(`
	type Student struct {
			ClassInfo
			Name      string
	}
	type DupStudent struct{
			ClassInfo
			Name      string
			No        int
	}

	var s1 DupStudent = DupStudent{
		ClassInfo:ClassInfo{Class:10, No:20},
		Name:"John",
		No:100,
	}
	`)
	

	var s1 DupStudent = DupStudent{
		ClassInfo:ClassInfo{Class:10, No:20},
		Name:"John",
		No:100,
	}
	fmt.Println("Class : ", s1.Class)
	fmt.Println("NO    : ", s1.No)
	fmt.Println("Name  : ", s1.Name)
	fmt.Println("ClassInfo.Class :", s1.ClassInfo.Class)
	fmt.Println("ClassInfo.No    :", s1.ClassInfo.No)
}
