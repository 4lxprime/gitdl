# gitdl
git downloader for specific folder

## cli:
you can run the cli with `go run github.com/4lxprime/gitdl@main -h` from anywhere on your system as long as you have go installed.

## code:
```golang
if err := gitdl.DownloadGit(
    "torvalds/linux",   // base the clone on https://github.com/torvalds/linux
    "kernel",           // only clone the folder "kernel" from linux
    "linux_qernel",     // name the folder "linux_qernel"
    gitdl.WithBranch("master"),                             // base the clone on the master branch (default main)
    gitdl.WithExclusions("configs/*", "async.c"),           // exclude everything in configs/ and exclude async.c
    gitdl.WithReplace(gitdl.Map{"linux", "linuxisbetter"})  // everytime "linux" is mentionned, it will be remplaced by "linuxisbetter"
    gitdl.WithLogs,                                         // you want to enable log
    gitdl.WithAuth("ghp_*")                                 // use your github auth token so you can have likely unlimited requests
    gitdl.WithoutChecksum                                   // don't care about security (not recommended) 
); err != nil {
    log.Fatal(err)
}
```
