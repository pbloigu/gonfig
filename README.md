# Configuration helper

Just a smaal piece of code that reads a configuration file in YAML format (as root) and the drops the privileges to the ones specified in the configuration file.

Just use `gonfig.Get(path string, into RunAsConfig)` tho read the configuration from the given path. The only requirement is that the struct into which the configuration is parsed implements `gonfig.RunAsConfig`so that we know which gid and uid to assume after reading the file. 

Nothing special really. 