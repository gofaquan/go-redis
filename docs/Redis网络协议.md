## Redis Serialization Protocol (RESP)

redis 客户端与服务端之间通信的协议与特定格式

- 正常回复

- 错误回复

- 整数

- 多行字符串

- 数组

#### 正常回复

以"+"开头，以"\r\n"结尾的字符串形式

```
+消息内容\r\n
+OK\r\n
```

#### 错误回复

以"."开头,以"\r\n"结尾的字符串形式

```
-Error message\r\n
```

#### 多行字符串

以"$"开头，后跟实际发送字节数,以"\r\n"结尾

```
"gofaquan.com"
$12\r\gofaquan.com\r\n
""                    (空字符)
$0\r\n\r\n
"gofaquan\r\n123"     (不会发生转义)
$14\r\ngofaquan\r\n123\r\n
```

#### 数组

以 "*" 开头，后跟成员个数

```
SET key value
*3\r\n$3\r\nSET\r\n$3\r\nkey\r\n$5\r\nvalue\r\n

SET => $3\r\nSET\r\n
key => $3\r\nkey\r\n
value => $5\r\nvalue\r\n
```

实现在