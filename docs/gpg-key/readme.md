**TODO: REFACTOR -We need a GPG key in our pipeline to sign the binary. Setup can be found here**

You can generate a **PGP private key** using **GnuPG (GPG)**, which is the most common tool for managing OpenPGP keys. Here's how to do it:

---

### **🔹 Step 1: Install GPG**
If you don’t have GPG installed, install it using:

- **Linux (Debian/Ubuntu)**
  ```bash
  sudo apt update && sudo apt install gnupg -y
  ```

- **MacOS (Homebrew)**
  ```bash
  brew install gnupg
  ```

- **Windows**
  - Download and install **Gpg4win** from: [https://gpg4win.org/download.html](https://gpg4win.org/download.html)

---

### **🔹 Step 2: Generate a New PGP Key**
Run the following command:

```bash
gpg --full-generate-key
```

You'll be asked a few questions:

1. **Choose the key type:**  
   - Select **1** (**RSA and RSA**) (Recommended)
   
2. **Key size:**  
   - Enter **4096** (Higher security)

3. **Expiration date:**  
   - Choose **0** (never expire) or set a custom duration

4. **User ID Information:**  
   - **Real Name**: e.g., "Michiel Van Haegenborgh"
   - **Email Address**: e.g., "your.email@example.com"
   - **Comment**: (Optional)

5. **Passphrase:**  
   - Choose a **strong passphrase** (you’ll need it for signing)

---

### **🔹 Step 3: Verify Your Key**
Run:

```bash
gpg --list-secret-keys --keyid-format=long
```

Example output:

```
sec   rsa4096/ABCDE1234567890 2025-03-13 [SC]
      Key fingerprint = 1234 5678 9ABC DEF0 1234 5678 9ABC DEF0 1234 5678
uid   Michiel Van Haegenborgh <your.email@example.com>
ssb   rsa4096/09876FEDCBA54321 2025-03-13 [E]
```

Take note of **your key ID** (`ABCDE1234567890` in this example).

---

### **🔹 Step 4: Export Your PGP Private Key**
To use this key in **GitHub Actions**, export it:

```bash
gpg --armor --export-secret-keys YOUR-KEY-ID > pgp-private-key.asc
```

Replace `YOUR-KEY-ID` with your actual key ID (e.g., `ABCDE1234567890`).

---

### **🔹 Step 5: Add the Key to GitHub Secrets**
1. Open **GitHub Repository → Settings → Secrets and Variables → Actions**  
2. Click **New repository secret**  
3. Name it **`PGP_PRIVATE_KEY`**  
4. Paste the contents of `pgp-private-key.asc`  

---

### **🔹 Step 6: Use the PGP Key in GitHub Actions**
Modify your **GitHub Actions workflow** to import the key before running GoReleaser:

```yaml
- name: Import PGP Key
  run: |
    echo "${{ secrets.PGP_PRIVATE_KEY }}" | gpg --import
  env:
    GPG_TTY: $(tty)
```

---

### **🔹 Step 7: Get the PGP Key Fingerprint**
Run:

```bash
gpg --list-secret-keys --keyid-format=long
```

You'll get something like:

```
pub   rsa4096/ABCDE1234567890 2025-03-13 [SC]
      Key fingerprint = 1234 5678 9ABC DEF0 1234 5678 9ABC DEF0 1234 5678
```

Copy **only the fingerprint** (e.g., `123456789ABCDEF0123456789ABCDEF012345678`) and add it to **GitHub Secrets** as:

- **`GPG_FINGERPRINT`**

Now, your GoReleaser config should work with signing! 🚀

Let me know if you need any clarifications. 🔐