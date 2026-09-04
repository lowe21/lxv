package configcenter

import (
	"sync"

	"dubbo.apache.org/dubbo-go/v3"
	"dubbo.apache.org/dubbo-go/v3/common/config"
	"dubbo.apache.org/dubbo-go/v3/common/constant"
	"dubbo.apache.org/dubbo-go/v3/config_center"
	"dubbo.apache.org/dubbo-go/v3/global"
	_ "dubbo.apache.org/dubbo-go/v3/imports"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcfg"
	"github.com/gogf/gf/v2/os/genv"
	"github.com/gogf/gf/v2/util/gconv"
)

var once sync.Once

func Init() {
	once.Do(func() {
		dubboConfigFile := genv.Get(constant.ConfigFileEnvKey).String()
		if dubboConfigFile != "" {
			configAdapter := g.Config().GetAdapter().(*gcfg.AdapterFile)
			configAdapter.SetFileName(dubboConfigFile)

			centerConfig := &global.CenterConfig{}
			if err := configAdapter.MustGet(nil, constant.ConfigCenterPrefix).Scan(centerConfig); err != nil {
				panic(err)
			}

			if _, err := dubbo.NewInstance(
				dubbo.WithConfigCenter(
					config_center.WithConfigCenter(centerConfig.Protocol),
					config_center.WithAddress(centerConfig.Address),
					config_center.WithDataID(centerConfig.DataId),
					config_center.WithCluster(centerConfig.Cluster),
					config_center.WithGroup(centerConfig.Group),
					config_center.WithUsername(centerConfig.Username),
					config_center.WithPassword(centerConfig.Password),
					config_center.WithNamespace(centerConfig.Namespace),
					config_center.WithAppID(centerConfig.AppID),
					config_center.WithTimeout(gconv.Duration(centerConfig.Timeout)),
					config_center.WithParams(centerConfig.Params),
					config_center.WithFileExtYaml(),
				),
			); err != nil {
				panic(err)
			}

			properties, err := config.GetEnvInstance().GetDynamicConfiguration().GetProperties(centerConfig.DataId, config_center.WithGroup(centerConfig.Group))
			if err != nil {
				panic(err)
			}

			configAdapter.SetContent(properties)
			configAdapter.SetFileName(gcfg.DefaultConfigFileName)

			if err = g.Server().SetConfigWithMap(configAdapter.MustGet(nil, "server").Map()); err != nil {
				panic(err)
			}
		} else {
			configAdapter := g.Config().GetAdapter().(*gcfg.AdapterFile)
			configFile, err := configAdapter.GetFilePath()
			if err != nil {
				panic(err)
			}

			if err = genv.Set(constant.ConfigFileEnvKey, configFile); err != nil {
				panic(err)
			}
		}
	})
}
