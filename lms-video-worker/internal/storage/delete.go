package storage
import ("os"; "strings")
func DeleteBlobFromUrl(url string) error {
 if !strings.HasPrefix(url,"/uploads/") { return nil }
 path,err:=pathFor(strings.TrimPrefix(url,"/uploads/")); if err!=nil { return err }; err=os.Remove(path); if os.IsNotExist(err) { return nil }; return err
}
