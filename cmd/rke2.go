/*
Copyright © 2025 EDGEFORGE contact@edgeforge.eu

Only supported on linux because bash dependencies and containers on windows.. yeah, nope.
*/
package cmd

import (
	"fmt"
	"os"
/* 	"strings"

	"github.com/google/uuid" */
	common "github.com/michielvha/edgectl/pkg/common"
	vault "github.com/michielvha/edgectl/pkg/vault"
	server	"github.com/michielvha/edgectl/pkg/rke2/server" // Import the new package
	"github.com/spf13/cobra"
)

// TODO: Move functions to a separate package. Only keep the cobra command logic here.
// TODO: Create function to store kubeconfig file in vault for later usage.

/* //go:embed scripts/*.sh
var embeddedScripts embed.FS

// Extracts an embedded script to /tmp
func extractEmbeddedScript(scriptName string) string {
	scriptPath := filepath.Join("/tmp", scriptName)

	// Read script from embedded FS
	data, err := embeddedScripts.ReadFile("scripts/" + scriptName)
	if err != nil {
		fmt.Printf("❌ Failed to read embedded script: %v\n", err)
		os.Exit(1)
	}

	// Write to a temp file
	if err := os.WriteFile(scriptPath, data, 0o755); err != nil {
		fmt.Printf("❌ Failed to write script: %v\n", err)
		os.Exit(1)
	}

	return scriptPath
}

// Runs a function from the sourced script
func runBashFunction(scriptName, functionName string) {
	scriptPath := extractEmbeddedScript(scriptName)

	// // Run the function from the sourced script
	// cmd := exec.Command("bash", "-c", fmt.Sprintf("source %s && %s", scriptPath, functionName))
	// cmd.Stdout = os.Stdout
	// cmd.Stderr = os.Stderr
	// if err := cmd.Run(); err != nil {
	// 	fmt.Printf("❌ Error executing function %s from %s: %v\n", functionName, scriptPath, err)
	// 	os.Exit(1)
	// }

	// Run the full script and pass the function name to call inside the script
	cmd := exec.Command("bash", scriptPath, functionName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin // Important to inherit input in case sudo or interactive steps exist
	if err := cmd.Run(); err != nil {
		fmt.Printf("❌ Error executing %s from %s: %v\n", functionName, scriptPath, err)
		os.Exit(1)
	}
}
*/
// rke2Cmd represents the "rke2" command
var rke2Cmd = &cobra.Command{
	Use:   "rke2",
	Short: "Manage RKE2 cluster",
	Long: `The "rke2" command allows you to install, manage, and uninstall RKE2.

Examples:
  edgectl rke2 server      # Install RKE2 Server
  edgectl rke2 agent       # Install RKE2 Agent
  edgectl rke2 uninstall   # Uninstall RKE2
`,
}

// Install RKE2 Server
// var installLoadBalancerCmd = &cobra.Command{
// 	Use:   "lb",
// 	Short: "Install RKE2 load balancer",
// 	Run: func(cmd *cobra.Command, args []string) {
// 		fmt.Println("🚀 Install a load balancer for RKE2...")
// 		runBashFunction("rke2.sh", "install_rke2_lb")
// 	},
// }

// Fetch token from Vault & set as env var / file
func fetchTokenFromVault(clusterID string) string {
	fmt.Println("🔐 Cluster ID supplied, retrieving join token from Vault...")

	vaultClient, err := vault.NewClient()
	if err != nil {
		fmt.Printf("❌ Failed to initialize Vault client: %v\n", err)
		os.Exit(1)
	}

	token, err := vaultClient.RetrieveJoinToken(clusterID)
	if err != nil {
		fmt.Printf("❌ Failed to retrieve join token from Vault: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Retrieved token: %s\n", token)
	_ = os.WriteFile("/etc/edgectl/cluster-id", []byte(clusterID), 0o644)
	_ = os.Setenv("RKE2_TOKEN", token)

	return token
}

var installServerCmd = &cobra.Command{
	Use:   "server install",
	Short: "Install RKE2 Server",
	Run: func(cmd *cobra.Command, args []string) {
		// ✅ Check if user is root
		if os.Geteuid() != 0 {
			fmt.Println("❌ This command must be run as root. Try using `sudo`.")
			os.Exit(1)
		}

		// ✅ Extract values and call logic
		clusterID, _ := cmd.Flags().GetString("cluster-id")
		isExisting := cmd.Flags().Changed("cluster-id")

		err := server.Install(clusterID, isExisting)
		if err != nil {
			fmt.Printf("❌ RKE2 server install failed: %v\n", err)
			os.Exit(1)
		}
	},
}

/* // Install RKE2 Server
var installServerCmd = &cobra.Command{
	Use:   "server",
	Short: "Install RKE2 Server",
	Run: func(cmd *cobra.Command, args []string) {
		// check if root, needed to upload file
		if os.Geteuid() != 0 {
			fmt.Println("❌ This command must be run as root. Try using `sudo`.")
			os.Exit(1)
		}

		fmt.Println("🚀 Installing RKE2 Server...")

		// Reuse our vault abstraction in ``pkg/vault/rke2-handler.go``
		vaultClient, err := vault.NewClient()
		if err != nil {
			fmt.Printf("❌ Failed to initialize Vault client: %v\n", err)
			os.Exit(1)
		}

		clusterID, _ := cmd.Flags().GetString("cluster-id")

		if clusterID != "" {
			// fetch the token from the vault
			fetchTokenFromVault(clusterID)
		} else {
			// if token is not supplied create it
			clusterID = fmt.Sprintf("rke2-%s", uuid.New().String()[:8])
			_ = os.WriteFile("/etc/edgectl/cluster-id", []byte(clusterID), 0o644)
			fmt.Printf("🆔 Generated cluster ID: %s\n", clusterID)
		}

		common.RunBashFunction("rke2.sh", "install_rke2_server")

		// Store the token & kubeconfig in vault if cluster-id wasn't supplied
		if !cmd.Flags().Changed("cluster-id") {
			tokenBytes, err := os.ReadFile("/var/lib/rancher/rke2/server/node-token")
			if err != nil {
				fmt.Printf("❌ Failed to read generated node token: %v\n", err)
				os.Exit(1)
			}

			token := strings.TrimSpace(string(tokenBytes))
			if err := vaultClient.StoreJoinToken(clusterID, token); err != nil {
				fmt.Printf("❌ Failed to store token in Vault: %v\n", err)
				os.Exit(1)
			}

			fmt.Printf("🔐 Token successfully stored in Vault for cluster %s\n", clusterID)

			// Check if kubeconfig exists
			// TODO: handle changing 127.0.0.1 to the load balancer IP
			kubeconfigPath := "/etc/rancher/rke2/rke2.yaml"
			if _, statErr := os.Stat(kubeconfigPath); os.IsNotExist(statErr) {
				fmt.Printf("❌ Kubeconfig file not found at path: %s\n", kubeconfigPath)
				os.Exit(1)
			}

			// if it exists store it vault
			err = vaultClient.StoreKubeConfig(clusterID, kubeconfigPath)
			if err != nil {
				fmt.Printf("❌ Failed to store kubeconfig in Vault: %v\n", err)
				os.Exit(1)
			} else {
				fmt.Printf("🔐 Kubeconfig successfully stored in Vault for cluster %s\n", clusterID)
			}
		}
	},
} */

// Install RKE2 Agent
var installAgentCmd = &cobra.Command{
	Use:   "agent",
	Short: "Install RKE2 Agent",
	Run: func(cmd *cobra.Command, args []string) {
		clusterID, _ := cmd.Flags().GetString("cluster-id")
		// cobra already checks this
		// if clusterID == "" {
		// 	fmt.Println("❌ cluster ID is required to join an existing cluster.")
		// 	os.Exit(1)
		// }
		fetchTokenFromVault(clusterID) // this will fetch the token and safe as env var to be used in bash function.
		// TODO: figure how to dynamically set lb hostname/ip as env var...
		common.RunBashFunction("rke2.sh", "install_rke2_agent -l 192.168.10.125")
	},
}

// Check RKE2 status
// TODO: Add more status checks
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show status of RKE2",
	Run: func(cmd *cobra.Command, args []string) {
		common.RunBashFunction("rke2.sh", "rke2_status")
	},
}

// Uninstall RKE2
var uninstallCmd = &cobra.Command{
	Use:   "purge",
	Short: "purge RKE2 install from host",
	Run: func(cmd *cobra.Command, args []string) {
		common.RunBashFunction("rke2.sh", "purge_rke2")
	},
}

// Configure kubeconfig
var SetKubeConfigCmd = &cobra.Command{
	Use:   "config kube",
	Short: "Fetch kubeconfig from Vault and store it on the host",
	Run: func(cmd *cobra.Command, args []string) {
		clusterID, _ := cmd.Flags().GetString("cluster-id")
		outputPath, _ := cmd.Flags().GetString("output")

		// Cobra already checks this when using `MarkFlagRequired`
		// if clusterID == "" {
		// 	fmt.Println("❌ You must provide a --cluster-id")
		// 	os.Exit(1)
		// }

		vaultClient, err := vault.NewClient()
		if err != nil {
			fmt.Printf("❌ Failed to initialize Vault client: %v\n", err)
			os.Exit(1)
		}

		err = vaultClient.RetrieveKubeConfig(clusterID, outputPath)
		if err != nil {
			fmt.Printf("❌ Failed to retrieve kubeconfig: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✅ Kubeconfig successfully written to: %s\n", outputPath)

		// Configure bash shell to use the kubeconfig
		common.RunBashFunction("rke2.sh", "configure_rke2_bash")
	},
}

// Register subcommands
func init() {
	// Attach rke2 as rootCmd
	rootCmd.AddCommand(rke2Cmd)

	// installServerCmd Flags
	installServerCmd.Flags().String("cluster-id", "", "The clusterID required to join an existing cluster")
	// installAgentCmd Flags
	installAgentCmd.Flags().String("cluster-id", "", "The ID of the cluster you want to join")
	_ = installAgentCmd.MarkFlagRequired("cluster-id")

	SetKubeConfigCmd.Flags().String("cluster-id", "", "The ID of the cluster to fetch the kubeconfig for")
	SetKubeConfigCmd.Flags().String("output", "/etc/rancher/rke2/rke2.yaml", "Destination path to store the kubeconfig")
	_ = SetKubeConfigCmd.MarkFlagRequired("cluster-id")

	// Attach subcommands under rke2
	rke2Cmd.AddCommand(installServerCmd)
	rke2Cmd.AddCommand(installAgentCmd)
	rke2Cmd.AddCommand(statusCmd)
	rke2Cmd.AddCommand(uninstallCmd)
	rke2Cmd.AddCommand(SetKubeConfigCmd)
}
