# Testing

> [!NOTE]
> There are barely any tests at the moment in the package, this is mostly future proofing.

To test ptxt, you need to place a `.ggfnt` font inside `test/fonts`. Once that's done, you can simply:
```
go test -tags cputext .
```
You can also test without the `-tags cputext` mode and use the default version with Ebitengine, but some tests will be omitted. The CPU mode is recommended while testing because it doesn't depend on Ebitengine.
