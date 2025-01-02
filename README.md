# Configuration helper

Just a small piece of code that reads a configuration file in YAML format (as root) and then drops the privileges to the ones specified in the configuration file.

Just use `gonfig.Get(path string, into RunAsConfig)` to read the configuration from the given path. The only requirement is that the struct into which the configuration is parsed implements `gonfig.RunAsConfig`so that we know which gid and uid to assume after reading the file. 

Nothing special really. 
