fix
```
func sql.NullStringSlice2interface(l []sql.NullString) []interface{} {
                   v := make([]interface{}, len(l))
                   for i, val := range l {
                           v[i] = val
           
                   }
                   return v
           }
```