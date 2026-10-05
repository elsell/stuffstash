package app

import (
	"context"
	"testing"
)

type requestInput []byte

func (b requestInput) Read(context.Context, string) ([]byte, error) { return b, nil }
func TestDirectoryBodyValidationBeforeWrite(t *testing.T) {
	for _, body := range []string{`[]`, `null`, `{"name":"x"} {}`, `{`} {
		r := Runner{InputFiles: requestInput(body)}
		if _, err := r.prepareInput(context.Background(), Options{Command: []string{"tenants", "create"}, InputPath: "-"}); err == nil {
			t.Fatalf("invalid body accepted: %s", body)
		}
	}
	body := `{"name":null}`
	r := Runner{InputFiles: requestInput(body)}
	got, err := r.prepareInput(context.Background(), Options{Command: []string{"tenants", "update"}, InputPath: "-"})
	if err != nil || string(got.RequestBody) != body {
		t.Fatalf("body changed: %s %v", got.RequestBody, err)
	}
	if _, err := r.prepareInput(context.Background(), Options{Command: []string{"assets", "list"}, InputPath: "-"}); err == nil {
		t.Fatal("ignored body accepted")
	}
}

func TestDirectoryWritesDoNotPromiseUnsupportedIdempotency(t *testing.T) {
	r := Runner{}
	if _, err := r.prepareInput(context.Background(), Options{Command: []string{"tenants", "create"}, ConnectorName: "Home", IdempotencyKey: "retry"}); err == nil {
		t.Fatal("unsupported retry key accepted")
	}
}

type namePrompt struct{calls int}
func(p *namePrompt)ReadText(context.Context,string,int)(string,error){p.calls++;return "Garage",nil}
func TestMissingNamePromptsOnlyInInteractiveMode(t *testing.T){
 prompt:=&namePrompt{}
 runner:=Runner{TextInput:prompt}
 options:=Options{Command:[]string{"inventories","create"}}
 result,err:=runner.prepareInput(context.Background(),options)
 if err!=nil || string(result.RequestBody)!=`{"name":"Garage"}` || prompt.calls!=1{t.Fatalf("prompt failed: %s %v",result.RequestBody,err)}
 for _,o:=range []Options{{Command:options.Command,JSON:true},{Command:options.Command,NoInput:true}}{
  if _,err:=runner.prepareInput(context.Background(),o);err==nil{t.Fatal("script accepted missing name")}
 }
 if prompt.calls!=1{t.Fatal("script opened prompt")}
}
