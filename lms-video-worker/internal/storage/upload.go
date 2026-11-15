package storage
import (
 "fmt"
 "io"
 "os"
 "path/filepath"
 "strings"
)
func pathFor(name string) (string,error) {
 name=filepath.Clean(name); if name=="." || filepath.IsAbs(name) || name==".." || strings.HasPrefix(name,"../") { return "",fmt.Errorf("invalid path") }
 root:=os.Getenv("STORAGE_DIR"); if root=="" { root="../.local/uploads" }; return filepath.Join(root,name),nil
}
func UploadFile(srcPath,name string) (string,error) {
 target,err:=pathFor(name); if err!=nil { return "",err }; if err=os.MkdirAll(filepath.Dir(target),0700); err!=nil { return "",err }
 src,err:=os.Open(srcPath); if err!=nil { return "",err }; defer src.Close()
 dst,err:=os.Create(target); if err!=nil { return "",err }; _,err=io.Copy(dst,src); closeErr:=dst.Close(); if err!=nil { return "",err }; if closeErr!=nil { return "",closeErr }
 return "/uploads/"+filepath.ToSlash(name),nil
}
