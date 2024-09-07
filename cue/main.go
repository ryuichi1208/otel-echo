package main

import (
	"fmt"
	"log"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/load"
)

// Person はGoの構造体で、CUEスキーマに準拠する必要があります。
type Person struct {
	Name string
	Age  int `cue:"<120"`
}

func main() {
	// CUEコンテキストを作成
	ctx := cuecontext.New()

	// CUEファイルをロード（現在のディレクトリから）
	instances := load.Instances([]string{"./cue.mod/gen/main/"}, nil)

	// インスタンスをビルド
	inst := ctx.BuildInstance(instances[0])
	if inst.Err() != nil {
		log.Fatalf("CUE build error: %v", inst.Err())
	}

	// CUEスキーマに基づいてGoの構造体を検証
	person := Person{"John Doe", 121}
	err := validatePerson(ctx, inst, person)
	if err != nil {
		log.Fatalf("Validation failed: %v", err)
	}

	fmt.Println("Validation succeeded:", person)
}

// validatePerson は、GoのPerson構造体をCUEスキーマに対して検証します。
func validatePerson(ctx *cue.Context, value cue.Value, p Person) error {
	// CUE値にエンコード
	personValue := ctx.Encode(p)

	// CUEスキーマを取得
	personCue := value.LookupDef("#Person")
	if personCue.Err() != nil {
		return personCue.Err()
	}

	// ユニフィケーションと検証
	unified := personCue.Unify(personValue)
	return unified.Validate()
}
