package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	guuid "github.com/google/uuid"
	"golang.org/x/crypto/ssh"
)

var keyForCryptPassPhrase = "YourSecretEncryptionKey!"

var connectionPool = make(map[string]*ConnectionContext)

var LastConnectedSSHConnectionID string

type ConnectionContext struct {
	Client  *ssh.Client
	Config  SSHConfig
	Session *ssh.Session // This could be added if you want to persist sessions too
}

// Define the request structure for SSH connection
type SSHRequest struct {
	ID            string `json:"id"`
	Host          string `json:"host"`
	Port          string `json:"port"`
	User          string `json:"user"`
	Password      string `json:"password"`
	PrivateKey    string `json:"privateKey"`
	KeyFilePath   string `json:"keyFilePath"`
	KeyPassphrase string `json:"keyPassphrase"`
	SaveParams    bool   `json:"saveParams"`
	Command       string `json:"command"`
}

type SSHConfig struct {
	Host          string
	Port          string
	User          string
	Password      string
	PrivateKey    []byte
	KeyPassphrase string
}

func establishSSHConnection(config *SSHConfig, sshReq SSHRequest) (string, error) {
	// config, err := generateSSHConfig(sshReq)
	// if err != nil {
	// 	return "", err
	// }

	client, err := ConnectToSSH(*config) // Assume this returns a client
	if err != nil {
		return "", err
	}

	// Generate a unique ID for this connection. This can be refined further.
	id := genUUID() // Assume you have this function
	connectionPool[id] = &ConnectionContext{
		Client: client,
		Config: *config,
	}

	LastConnectedSSHConnectionID = id

	fmt.Printf("Connection established with ID: %s\n", id)

	return id, nil
}

func closeSSHConnection(id string) error {
	context, exists := connectionPool[id]
	if !exists {
		return fmt.Errorf("Connection not found")
	}

	err := context.Client.Close()
	if err != nil {
		return err
	}

	delete(connectionPool, id)
	return nil
}

func sshCloseHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	err := closeSSHConnection(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func executeSSHCommandHandler(w http.ResponseWriter, r *http.Request) {

	sshReq, err := readSSHRequestFromBody(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	fmt.Printf("Executing command: %s on connection: %s\n", sshReq.Command, sshReq.ID)
	context, exists := connectionPool[sshReq.ID]
	if !exists {
		http.Error(w, "Connection not found", http.StatusNotFound)
		return
	}

	// Use context.Client for further operations...
	out, err := ExecuteSSHCommand(sshReq.Command, context.Client)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Output the result
	w.Write([]byte(out))
}

func sshConnectHandler(w http.ResponseWriter, r *http.Request) {
	sshReq, err := readSSHRequestFromBody(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	config, err := generateSSHConfig(sshReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = handlePassword(&config, sshReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = handlePassphrase(&config, sshReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	id, err := establishSSHConnection(&config, sshReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write([]byte(id))
}

func sshListActiveConnectionsHandler(w http.ResponseWriter, r *http.Request) {

	// out JSON array of active connections

	out := []string{}
	for id, _ := range connectionPool {
		out = append(out, id)
	}

	json.NewEncoder(w).Encode(out)
}

func readSSHRequestFromBody(r *http.Request) (SSHRequest, error) {
	var sshReq SSHRequest
	reqBody, err := io.ReadAll(r.Body)
	if err != nil {
		return sshReq, err
	}

	err = json.Unmarshal(reqBody, &sshReq)
	return sshReq, err
}

func generateSSHConfig(sshReq SSHRequest) (SSHConfig, error) {
	var keyContent []byte
	var err error
	if sshReq.KeyFilePath != "" {
		keyContent, err = os.ReadFile(sshReq.KeyFilePath)
		if err != nil {
			return SSHConfig{}, err
		}
	}

	config := SSHConfig{
		Host:          sshReq.Host,
		Port:          sshReq.Port,
		User:          sshReq.User,
		Password:      sshReq.Password,
		PrivateKey:    keyContent,
		KeyPassphrase: sshReq.KeyPassphrase,
	}
	return config, nil
}

func handlePassphrase(config *SSHConfig, sshReq SSHRequest) error {
	if sshReq.KeyFilePath != "" {
		passphraseFilePath := TruncateSHA256Hash(config.PrivateKey, 12) + ".secret"
		// Handle passphrase
		if sshReq.SaveParams {
			if _, err := os.Stat(passphraseFilePath); !os.IsNotExist(err) {

				encryptedData, err := os.ReadFile(passphraseFilePath)
				if err != nil {
					return fmt.Errorf("Failed to read encrypted passphrase file: %v", err)
				}

				decryptedData, err := decrypt(encryptedData, keyForCryptPassPhrase)
				if err != nil {
					return fmt.Errorf("Failed to decrypt passphrase: %v", err)
				}
				if sshReq.KeyPassphrase == "" {
					config.KeyPassphrase = string(decryptedData)
				}

			} else {
				if sshReq.KeyPassphrase != "" {
					encryptedPassphrase, err := encrypt([]byte(sshReq.KeyPassphrase), keyForCryptPassPhrase)
					if err != nil {
						return fmt.Errorf("Failed to encrypt passphrase: %v", err)
					}

					err = os.WriteFile(passphraseFilePath, encryptedPassphrase, 0644)
					if err != nil {
						return fmt.Errorf("Failed to write to passphrase file: %v", err)
					}
				}
			}
		} else {
			if _, err := os.Stat(passphraseFilePath); !os.IsNotExist(err) {
				err := os.Remove(passphraseFilePath)
				if err != nil {
					return fmt.Errorf("Failed to delete passphrase file: %v", err)
				}
			}
		}
	}

	return nil
}

func handlePassword(config *SSHConfig, sshReq SSHRequest) error {
	if sshReq.Password == "" {
		passwordFilePath := TruncateSHA256Hash([]byte(sshReq.Host), 12) + ".secret"

		// Handle passphrase
		if sshReq.SaveParams {
			if _, err := os.Stat(passwordFilePath); !os.IsNotExist(err) {
				encryptedData, err := os.ReadFile(passwordFilePath)
				if err != nil {
					return fmt.Errorf("Failed to read encrypted passphrase file: %v", err)
				}

				decryptedData, err := decrypt(encryptedData, keyForCryptPassPhrase)
				if err != nil {
					return fmt.Errorf("Failed to decrypt passphrase: %v", err)
				}
				if sshReq.Password == "" {
					config.Password = string(decryptedData)
				}
			} else {
				if sshReq.Password != "" {
					encryptedPassword, err := encrypt([]byte(sshReq.Password), keyForCryptPassPhrase)
					if err != nil {
						return fmt.Errorf("Failed to encrypt passphrase: %v", err)
					}

					err = os.WriteFile(passwordFilePath, encryptedPassword, 0644)
					if err != nil {
						return fmt.Errorf("Failed to write to passphrase file: %v", err)
					}
				}
			}
		} else {
			if _, err := os.Stat(passwordFilePath); !os.IsNotExist(err) {
				err := os.Remove(passwordFilePath)
				if err != nil {
					return fmt.Errorf("Failed to delete passphrase file: %v", err)
				}
			}
		}
	}

	return nil
}

func handleSSHConnection(config SSHConfig, sshReq SSHRequest) (string, error) {
	client, err := ConnectToSSH(config)
	if err != nil {
		return "", err
	}
	defer client.Close()

	output, err := ExecuteSSHCommand(sshReq.Command, client)
	return output, err
}

// Inside main()
func main() {

	http.HandleFunc("/top", topHandler)
	http.HandleFunc("/tail-log", tailLogHandler)
	http.HandleFunc("/sshconnect", sshConnectHandler)
	http.HandleFunc("/sshclose", sshCloseHandler)
	http.HandleFunc("/sshcommand", executeSSHCommandHandler)
	http.HandleFunc("/sshlist", sshListActiveConnectionsHandler)
	http.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "pong")
	})

	http.ListenAndServe(":8080", nil)
}

func SHA256Hash(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

func TruncateSHA256Hash(data []byte, length int) string {
	hash := SHA256Hash(data)
	if length > len(hash) {
		length = len(hash)
	}
	return hash[:length]
}

func genUUID() string {
	id := guuid.New()
	return id.String()
}
