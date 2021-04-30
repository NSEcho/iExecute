# iExecute
One shot execute on iOS device. 

iExecute will spawn iproxy on random port in the range of 50000-60000. 
Then it will connect to that port, run the command passed on the command line and return the output.

# Configuration

You need to have `.iExecute` inside your home directory which should look something like this:

```yaml
username: root
rport:    22
```

Each time `iExecute` is run, it will read this config file and connect to the device.

# Running

```bash
$ cat ~/.iExecute
username: root
rport:    22
$ iExecute whoami
root
```
