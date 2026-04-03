** Step 1: build binary **  

```
go build -o btrack main.go  
```

** Step 2: Install Globally**  

```
# Option A: To ~/go/bin (recommended)
mkdir -p ~/go/bin
cp btrack ~/go/bin/

# Option B: To /usr/local/bin (system-wide, needs sudo)
sudo cp btrack /usr/local/bin/
```

** Step 3: Add Path **   

Edit Shell profile (~/.zshrc, ~/.bashrc etc)  
```
export PATH="$HOME/go/bin:$PATH"  
```

Reload Shell  
```
source ~/.bashrc  # or ~/.zshrc  
```

