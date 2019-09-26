package env // import "github.com/PremiereGlobal/stim/stimpacks/env"

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/PremiereGlobal/stim/stim"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

//Env base struct for processing environment stuff
type Env struct {
	name        string
	stim        *stim.Stim
	kubeVersion string
}

//New creates a new Env struct to use
func New() *Env {
	return &Env{name: "env"}
}

//Name returns the name of this stimPack
func (v *Env) Name() string {
	return v.name
}

//BindStim binds in stim for this object
func (v *Env) BindStim(s *stim.Stim) {
	v.stim = s
}

func (v *Env) Command(viper *viper.Viper) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "env shell <environment name>",
		Short: "Configures/Sets up for the said environment",
		Long:  `This allows you to quickly switch between aws/kubernetes/helm/kops environments`,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	var shellCmd = &cobra.Command{
		Use:   "shell",
		Short: "",
		Long:  ``,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	var initCmd = &cobra.Command{
		Use:   "init",
		Short: "",
		Long:  ``,
		Run: func(cmd *cobra.Command, args []string) {
			v.makeEnvDir()
			envs := viper.GetStringMap("env")
			for k, val := range envs {
				v.makeEnvNameDir(k)
				t := val.(map[string]interface{})
				v.stim.GetLogger().Debug("Initializing Environment:  ", k)
				for tk, tv := range t {
					if tk == "kubectl" {
						kd := &KubeDownloader{env: *v, version: tv.(string)}
						err := DownloadPackage(kd)
						if err != nil {
							v.stim.GetLogger().Fatal("Problem installing env:", k, "kubeversion:", tk, err)
						} else {
							v.stim.GetLogger().Debug("Installed env: ", k, "kubeversion:", tk)
						}
						MakeEnvLink(kd, v.GetEnvNameDir(k), k)
					} else if tk == "vault" {
						vd := &VaultDownloader{env: *v, version: tv.(string)}
						err := DownloadPackage(vd)
						if err != nil {
							v.stim.GetLogger().Warn("Problem installing env:", k, "kubeversion:", tk, err)
						} else {
							v.stim.GetLogger().Debug("Installed env: ", k, "kubeversion:", tk)
						}
						MakeEnvLink(vd, v.GetEnvNameDir(k), k)
					}
				}
			}
		},
	}

	var getCmd = &cobra.Command{
		Use:   "get",
		Short: "",
		Long:  ``,
		Run: func(cmd *cobra.Command, args []string) {
			var dlr Downloader
			v.makeEnvDir()
			kv := v.stim.ConfigGetString("kubectl")
			if kv != "" {
				if !strings.HasPrefix(kv, "v") {
					nv := "v" + kv
					v.stim.GetLogger().Info("Kube version set as:'{}', it MUST start with a 'v', changing to `{}`", kv, nv)
					kv = nv
				}
				dlr = &KubeDownloader{env: *v, version: kv}
			}
			vv := v.stim.ConfigGetString("vault")
			if vv != "" {
				if strings.HasPrefix(vv, "v") {
					nv := vv[1:]
					v.stim.GetLogger().Info("Vault version set as:'{}', it DoesNot start with a 'v', changing to `{}`", vv, nv)
					vv = nv
				}
				dlr = &VaultDownloader{env: *v, version: vv}
			}
			DownloadPackage(dlr)
		},
	}
	getCmd.PersistentFlags().String("kubectl", "", "Version of kubectl to get")
	viper.BindPFlag("kubectl", getCmd.PersistentFlags().Lookup("kubectl"))
	getCmd.PersistentFlags().String("vault", "", "Version of vault to get")
	viper.BindPFlag("vault", getCmd.PersistentFlags().Lookup("vault"))

	v.stim.BindCommand(shellCmd, cmd)
	v.stim.BindCommand(initCmd, cmd)
	v.stim.BindCommand(getCmd, cmd)

	return cmd
}

func (v *Env) GetEnvDir() string {
	cp, err := v.stim.ConfigGetStimConfigDir()
	if err != nil {
		v.stim.GetLogger().Fatal("Could not find config file path:{}", err)
	}
	return filepath.FromSlash(filepath.Join(cp, "/env"))
}

func (v *Env) GetEnvBinDir() string {
	return filepath.FromSlash(v.GetEnvDir() + "/bin")
}

func (v *Env) GetEnvNameDir(envName string) string {
	return filepath.FromSlash(v.GetEnvDir() + "/" + envName)
}

func (v *Env) makeEnvNameDir(envName string) (string, error) {
	envPath := v.GetEnvNameDir(envName)
	if _, err := os.Stat(envPath); os.IsNotExist(err) {
		v.stim.GetLogger().Debug("Creating dir: ", envPath)
		err := os.MkdirAll(envPath, 0755)
		if err != nil {
			return "", err
		}
	}
	return envPath, nil
}

func (v *Env) makeEnvDir() error {
	path := v.GetEnvBinDir()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		v.stim.GetLogger().Debug("Creating dir: ", path)
		err := os.MkdirAll(path, 0755)
		if err != nil {
			return err
		}
	}
	return nil
}
