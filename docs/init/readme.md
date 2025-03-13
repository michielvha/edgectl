Starting simple with Cobra and setting up a modular CLI foundation is the right way to go. Let's get you started with the basics of `edgectl`.

## Initial Setup

### **Step 1: Initialize the Cobra CLI Project**

in the repository root run

1. Install Cobra:
   ```bash
   go install github.com/spf13/cobra-cli@latest
   ```

2. Create a new Go project:
   ```bash
   go mod init github.com/michielvha/edge-cli
   ```

3. Initialize Cobra:
   ```bash
   cobra-cli init
   ```

4. Add a version command:
   ```bash
   cobra-cli add version
   ```

5. Edit `cmd/version.go` to print the version:
   ```go
   TODO copy from code
   ```

6. Build and run:
   ```bash
   go build -ldflags "-X 'github.com/michielvha/edge-cli/cmd.Version=0.1.0'" -o edgectl.exe
   ./edgectl.exe version
   ```

At this point, you should have a working CLI that prints the version. Once this is set up, we can move on to:
- Adding commands to call Bash scripts for admin tasks
- Integrating HashiCorp Vault
- Fetching kubeconfig automatically
- Bootstrapping ArgoCD