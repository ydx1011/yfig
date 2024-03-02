# fig

fig是一个轻量化的配置读取工具

## 安装
```
go get github.com/xfali/fig
```

## 使用
### 加载配置内容
```
config := fig.New()
err := config.ReadValue(strings.NewReader(test_ctx_str))
if err != nil {
    b.Fatal(err)
}
```
或者
```
config, err := fig.LoadJsonFile("config.json")
if err != nil {
    t.Fatal(err)
}
config, err := fig.LoadYamlFile("config.yaml")
if err != nil {
    t.Fatal(err)
}
```
### 通过key获取属性值（字符串）
```
v := config.Get("DataSources.default.DriverName", "")
```
### 通过key获得反序列化对象
```
port := 0
err = config.GetValue("ServerPort", &port)
```
## 读取环境变量
使用模板函数env读取环境变量:
* 如果env参数为1个，如环境变量不存在则返回错误
* 如果env参数为2个，如环境变量不存在则赋值为第二个参数（默认值）
```
DataSources:
  default:
    DriverNameGet0: "{{ env "CONTEXT_TEST_ENV" }}"
    DriverNameGet1: "{{ env "CONTEXT_TEST_ENV" "func1_return" }}"
```

也可以在配置中使用{{.Env.ENV_NAME}}来配置读取环境变量值，fig在加载时自动使用ENV_NAME的值替换相应内容：
* 如环境变量不存在则返回错误
```
DataSources:
  default:
    DriverName: "{{.Env.CONTEXT_TEST_ENV}}"
```

## 工具方法
|  方法   | 说明  |
|  :----  | :----  |
| fig.GetString  | 获得string类型属性值 |
| fig.GetBool  | 获得bool类型属性值 |
| fig.GetInt  | 获得int类型属性值 |
| fig.GetUint  | 获得uint类型属性值 |
| fig.GetInt64  | 获得int64类型属性值 |
| fig.GetUint64  | 获得uint64类型属性值 |
| fig.GetFloat32  | 获得float32类型属性值 |
| fig.GetFloat64  | 获得float64类型属性值 |

用法：
```
config, _ := fig.LoadJsonFile("config.json")

v := fig.GetBool(config)("LogResponse", false)

floatValue := fig.GetFloat32(config)("Value.float", 0)
```

## tag
### 属性值tag
fig提供直接填充struct的field的方法，使用tag:"fig"来标识属性名称：
```
type TestStruct struct {
	dummy1      int
	Port        int  `fig:"ServerPort"`
	LogResponse bool `fig:"LogResponse"`
	dummy2      int
	FloatValue  float32 `fig:"Value.float"`
	DriverName  string  `fig:"DataSources.default.DriverName"`
	dummy3      int
}
```
### 属性前缀tag
可以使用tag:"figPx"表明属性的前缀，在此之后的所有fig tag都会自动增加此前缀：
```
type TestStruct2 struct {
	x           string `figPx:"DataSources.default"`
	MaxIdleConn int
	DvrName     string `fig:"DriverName"`
}
```
使用fig.Fill方法根据tag填充struct：
```
config, _ := fig.LoadJsonFile("config.json")
test := TestStruct{}
err := fig.Fill(config, &test)
t.log(test)
```

## 使用限制
目前不允许使用包含“-”的名称作为field，否则无法正常解析（请使用下划线“_”代替）。