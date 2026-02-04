package main

import (
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
)

const reservedInitialKey = "initial"

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	cmd := os.Args[1]
	var err error

	switch cmd {
	case "login":
		err = cmdLogin()
	case "logout":
		err = cmdLogout()
	case "list":
		err = cmdList()
	case "add":
		err = cmdAdd(os.Args[2:])
	case "get":
		err = cmdGet(os.Args[2:])
	case "update":
		err = cmdUpdate(os.Args[2:])
	case "remove":
		err = cmdRemove(os.Args[2:])
	case "generate-uuid":
		err = cmdGenerate(os.Args[2:], "uuid")
	case "generate-64hex":
		err = cmdGenerate(os.Args[2:], "64hex")
	default:
		usage()
		os.Exit(1)
	}

	if err != nil {
		printError(err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "Usage:")
	fmt.Fprintln(os.Stderr, "  secled login")
	fmt.Fprintln(os.Stderr, "  secled logout")
	fmt.Fprintln(os.Stderr, "  secled list")
	fmt.Fprintln(os.Stderr, "  secled add <key>")
	fmt.Fprintln(os.Stderr, "  secled get <key>")
	fmt.Fprintln(os.Stderr, "  secled update <key>")
	fmt.Fprintln(os.Stderr, "  secled remove <key>")
	fmt.Fprintln(os.Stderr, "  secled generate-uuid <key>")
	fmt.Fprintln(os.Stderr, "  secled generate-64hex <key>")
}

func printError(err error) {
	msg := err.Error()
	if !strings.HasPrefix(msg, "Error:") {
		msg = "Error: " + msg
	}
	fmt.Fprintln(os.Stderr, msg)
}

func cmdLogin() error {
	password, err := readPassword("Master password: ")
	if err != nil {
		return err
	}
	if len(password) < 8 {
		fmt.Fprintln(os.Stderr, "Warning: password length is less than 8 characters")
	}

	path, err := ledgerPath()
	if err != nil {
		return err
	}

	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		params, err := defaultKDFParams()
		if err != nil {
			return err
		}
		led := newLedger(params)

		masterKey := deriveKey(password, led.Params)
		initEntry, err := encryptEntry(masterKey, reservedInitialKey, []byte(initialValue()))
		if err != nil {
			return err
		}
		led.Entries[reservedInitialKey] = initEntry

		if err := saveLedger(path, led); err != nil {
			return err
		}

		fmt.Fprintln(os.Stderr, "Ledger created:", path)
	} else if err == nil {
		led, err := loadLedger(path)
		if err != nil {
			return err
		}
		if _, err := verifyPassword(led, password); err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	fmt.Fprintln(os.Stdout, formatSetEnv(password))
	return nil
}

func cmdLogout() error {
	fmt.Fprintln(os.Stdout, formatUnsetEnv())
	return nil
}

func cmdList() error {
	path, err := ledgerPath()
	if err != nil {
		return err
	}
	led, err := loadLedger(path)
	if err != nil {
		return err
	}

	keys := make([]string, 0, len(led.Entries))
	for k := range led.Entries {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		fmt.Fprintln(os.Stdout, k)
	}
	return nil
}

func cmdAdd(args []string) error {
	key, err := parseKeyArg(args)
	if err != nil {
		return err
	}
	if key == reservedInitialKey {
		return errors.New("key 'initial' is reserved")
	}

	password, err := requirePassword()
	if err != nil {
		return err
	}

	path, err := ledgerPath()
	if err != nil {
		return err
	}
	led, err := loadLedger(path)
	if err != nil {
		return err
	}
	masterKey, err := verifyPassword(led, password)
	if err != nil {
		return err
	}

	if _, exists := led.Entries[key]; exists {
		return errors.New("key already exists (use update)")
	}

	secret, err := readSecret("Secret value: ")
	if err != nil {
		return err
	}

	enc, err := encryptEntry(masterKey, key, secret)
	if err != nil {
		return err
	}
	led.Entries[key] = enc

	if err := saveLedger(path, led); err != nil {
		return err
	}

	return nil
}

func cmdGet(args []string) error {
	key, err := parseKeyArg(args)
	if err != nil {
		return err
	}

	password, err := requirePassword()
	if err != nil {
		return err
	}

	path, err := ledgerPath()
	if err != nil {
		return err
	}
	led, err := loadLedger(path)
	if err != nil {
		return err
	}
	masterKey, err := verifyPassword(led, password)
	if err != nil {
		return err
	}

	e, ok := led.Entries[key]
	if !ok {
		return errors.New("key not found")
	}

	plaintext, err := decryptEntry(masterKey, key, e)
	if err != nil {
		return errors.New("invalid password or corrupted entry")
	}

	_, err = os.Stdout.Write(plaintext)
	return err
}

func cmdUpdate(args []string) error {
	key, err := parseKeyArg(args)
	if err != nil {
		return err
	}
	if key == reservedInitialKey {
		return errors.New("key 'initial' is reserved")
	}

	password, err := requirePassword()
	if err != nil {
		return err
	}

	path, err := ledgerPath()
	if err != nil {
		return err
	}
	led, err := loadLedger(path)
	if err != nil {
		return err
	}
	masterKey, err := verifyPassword(led, password)
	if err != nil {
		return err
	}

	if _, exists := led.Entries[key]; !exists {
		return errors.New("key not found")
	}

	secret, err := readSecret("New secret value: ")
	if err != nil {
		return err
	}

	enc, err := encryptEntry(masterKey, key, secret)
	if err != nil {
		return err
	}
	led.Entries[key] = enc

	return saveLedger(path, led)
}

func cmdRemove(args []string) error {
	key, err := parseKeyArg(args)
	if err != nil {
		return err
	}
	if key == reservedInitialKey {
		return errors.New("key 'initial' is reserved")
	}

	password, err := requirePassword()
	if err != nil {
		return err
	}

	path, err := ledgerPath()
	if err != nil {
		return err
	}
	led, err := loadLedger(path)
	if err != nil {
		return err
	}
	if _, err := verifyPassword(led, password); err != nil {
		return err
	}

	if _, exists := led.Entries[key]; !exists {
		return errors.New("key not found")
	}
	delete(led.Entries, key)

	return saveLedger(path, led)
}

func cmdGenerate(args []string, kind string) error {
	key, err := parseKeyArg(args)
	if err != nil {
		return err
	}
	if key == reservedInitialKey {
		return errors.New("key 'initial' is reserved")
	}

	password, err := requirePassword()
	if err != nil {
		return err
	}

	path, err := ledgerPath()
	if err != nil {
		return err
	}
	led, err := loadLedger(path)
	if err != nil {
		return err
	}
	masterKey, err := verifyPassword(led, password)
	if err != nil {
		return err
	}

	if _, exists := led.Entries[key]; exists {
		return errors.New("key already exists (use update)")
	}

	var value string
	switch kind {
	case "uuid":
		value, err = generateUUIDv4()
	case "64hex":
		value, err = generate64Hex()
	default:
		return errors.New("unknown generator")
	}
	if err != nil {
		return err
	}

	enc, err := encryptEntry(masterKey, key, []byte(value))
	if err != nil {
		return err
	}
	led.Entries[key] = enc

	return saveLedger(path, led)
}

func parseKeyArg(args []string) (string, error) {
	if len(args) == 0 {
		return "", errors.New("missing key")
	}
	if len(args) > 1 {
		return "", errors.New("key must be a single argument (use quotes for spaces)")
	}
	return args[0], nil
}

func requirePassword() (string, error) {
	password := os.Getenv("SECLED_MASTER")
	if strings.TrimSpace(password) == "" {
		return "", errors.New("Error: login required (SECLED_MASTER is not set)")
	}
	return password, nil
}

func verifyPassword(led *ledger, password string) ([]byte, error) {
	masterKey := deriveKey(password, led.Params)
	initEntry, ok := led.Entries[reservedInitialKey]
	if !ok {
		return nil, errors.New("missing initial entry in ledger")
	}
	if _, err := decryptEntry(masterKey, reservedInitialKey, initEntry); err != nil {
		return nil, errors.New("invalid password or corrupted ledger")
	}
	return masterKey, nil
}
