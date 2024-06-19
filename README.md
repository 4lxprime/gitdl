# gitdl
 git downloader for specific folder

 you can use it like this in your code:
```golang
if err := gitdl.DownloadGit(
    "torvalds/linux",   // base the clone on https://github.com/torvalds/linux
    "kernel",           // only clone the folder "kernel" from linux
    "linux_qernel",     // name the folder "linux_qernel"
    gitdl.WithBranch("master"),                             // base the clone on the master branch (default main)
    gitdl.WithExclusions("configs/*", "async.c"),           // exclude everything in configs/ and exclude async.c
    gitdl.WithReplace(gitdl.Map{"linux", "linuxisbetter"})  // everytime "linux" is mentionned, it will be remplaced by "linuxisbetter"
); err != nil {
    log.Fatal(err)
}
```

 or you can just use the cli tool with `go run cmd/gitdl/main.go -repo=torvalds/linux -folder=kernel -output=linux_kernel -logs` but there is no every options.
