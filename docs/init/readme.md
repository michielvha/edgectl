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
   go mod init github.com/michielvha/edgectl
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
   go build -ldflags "-X main.Version=1.2.3 -X main.Commit=abcd1234" -o edgectl.exe
   ./edgectl.exe version
   ```

7. On a remote host you can now run:
   ```bash
   go install github.com/michielvha/edgectl@latest
   ```