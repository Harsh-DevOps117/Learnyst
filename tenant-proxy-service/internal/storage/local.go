package storage
import (
 "context"
 "fmt"
 "io"
 "mime"
 "os"
 "path/filepath"
 "strings"
)
func InitStorage() {}
func FetchFile(ctx context.Context,slug,path string) (io.ReadCloser,*string,error) {
 if path=="" || path=="/" { path="index.html" }
 if slug=="" || strings.ContainsAny(slug,"/\\") || slug==".." { return nil,nil,fmt.Errorf("invalid academy slug") }
 clean:=filepath.Clean(path); if filepath.IsAbs(clean) || clean==".." || strings.HasPrefix(clean,"../") { return nil,nil,fmt.Errorf("invalid path") }
 root:=os.Getenv("SITES_DIR"); if root=="" { root="../.local/sites" }
 f,err:=os.Open(filepath.Join(root,slug,clean)); if err!=nil { return nil,nil,err }
 contentType:=mime.TypeByExtension(filepath.Ext(path)); return f,&contentType,nil
}
