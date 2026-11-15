package storage
import (
 "fmt"
 "io"
 "os"
 "path/filepath"
 "strings"
)
func InitStorage() {}
func UploadDirectory(tenantSlug,localDir string) error {
 if tenantSlug=="" || strings.ContainsAny(tenantSlug,"/\\") || tenantSlug==".." { return fmt.Errorf("invalid academy slug") }
 root:=os.Getenv("SITES_DIR"); if root=="" { root="../.local/sites" }
 return filepath.WalkDir(localDir,func(path string,d os.DirEntry,err error) error {
  if err!=nil { return err }; if d.Type()&os.ModeSymlink!=0 { return fmt.Errorf("symlinks are not supported") }
  rel,err:=filepath.Rel(localDir,path); if err!=nil { return err }; target:=filepath.Join(root,tenantSlug,rel)
  if d.IsDir() { return os.MkdirAll(target,0755) }
  src,err:=os.Open(path); if err!=nil { return err }; defer src.Close()
  dst,err:=os.Create(target); if err!=nil { return err }; _,err=io.Copy(dst,src); closeErr:=dst.Close(); if err!=nil { return err }; return closeErr
 })
}
