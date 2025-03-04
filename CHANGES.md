# Changelog

## [v2.17.0][v2.17.0] - 05 Mar 2025

* Rename `Match.Index` to `Submatch` ([gh-252][gh-252])
* Rename `Match.Name` to `NamedSubmatch` ([gh-252][gh-252])
* Rename `Match.Values`/`NotValues` to `HasSubmatches`/`NotHasSubmatches` ([gh-252][gh-252])
* Rename `Cookie.HasMaxAge`/`NotContainsMaxAge` to `ContainsMaxAge`/`NotContainsMaxAge` ([gh-252][gh-252])
* Add `Response.Reader` for accessing response body directly ([gh-382][gh-382])
* Add `Request.WithRetryPolicyFunc` ([gh-435][gh-435])
* Add `Request.WithQueryEncoder` and `QueryEncoderFormKeepZeros` ([gh-438][gh-438])
* Add `Number.InDeltaRelative` ([gh-306][gh-306])
* Add `Object.Length`
* Update `Request.WithPath` to always format floats in decimal notation ([gh-449][gh-449])
* Fix panic in `DebugPrinter` when printing body ([gh-444][gh-444])
* Fix bugs in `NewMatchC()` and `NewWebsocketMessageC()`
* Improve test coverage
* Improve documentation and examples
* Add TLS example ([gh-205][gh-205])
* Bump some dependencies

[v2.17.0]: https://github.com/gavv/httpexpect/releases/tag/v2.17.0

[gh-205]: https://github.com/gavv/httpexpect/issues/205
[gh-252]: https://github.com/gavv/httpexpect/issues/252
[gh-306]: https://github.com/gavv/httpexpect/issues/306
[gh-382]: https://github.com/gavv/httpexpect/issues/382
[gh-435]: https://github.com/gavv/httpexpect/issues/435
[gh-438]: https://github.com/gavv/httpexpect/issues/438
[gh-444]: https://github.com/gavv/httpexpect/issues/444
[gh-449]: https://github.com/gavv/httpexpect/issues/449

## [v2.16.0][v2.16.0] - 03 Oct 2023

* Bump minimum Go version to 1.19
* Rename `Response.ContentType`/`ContentEncoding`/`TransferEncoding` to `HasXxx` ([gh-252][gh-252])
* Add `Request.WithReporter` and `WithAssertionHandler` ([gh-234][gh-234])
* Add stacktrace printing support ([gh-160][gh-160])
* Colorize JSON values ([gh-334][gh-334])
* Colorize HTTP requests and responses ([gh-343][gh-343])
* Prevent panic if `flag.Parse` hasn't been called ([gh-410][gh-410])
* Refactor and cleanup tests
* Improve test coverage
* Improve documentation
* Improve CI

[v2.16.0]: https://github.com/gavv/httpexpect/releases/tag/v2.16.0

[gh-160]: https://github.com/gavv/httpexpect/issues/160
[gh-234]: https://github.com/gavv/httpexpect/issues/234
[gh-252]: https://github.com/gavv/httpexpect/issues/252
[gh-334]: https://github.com/gavv/httpexpect/issues/334
[gh-343]: https://github.com/gavv/httpexpect/issues/343
[gh-410]: https://github.com/gavv/httpexpect/issues/410

## [v2.15.0][v2.15.0] - 04 Apr 2023

* Bump minimal Go version to 1.17
* Bump golang.org/x/net to 0.7.0
* Support colored output ([gh-161][gh-161], ([gh-335][gh-335]))
* Support thousand separation when formatting numbers ([gh-274][gh-274])
* Print HTTP request and response in failure message ([gh-159][gh-159])
* Add `RequestFactoryFunc`, `ClientFunc`, `WebsocketDialerFunc`, `ReporterFunc`, `LoggerFunc` ([gh-249][gh-249])
* Add `PanicReporter` ([gh-248][gh-248])
* Fix error messages in `AssertReporter` and `RequireReporter` (1e30c2736042d3b7e88bfff1948b74af4e8f1306)
* Update documentation
* Improve test coverage
* Refactor and cleanup tests

[v2.15.0]: https://github.com/gavv/httpexpect/releases/tag/v2.15.0

[gh-159]: https://github.com/gavv/httpexpect/issues/159
[gh-161]: https://github.com/gavv/httpexpect/issues/161
[gh-248]: https://github.com/gavv/httpexpect/issues/248
[gh-249]: https://github.com/gavv/httpexpect/issues/249
[gh-274]: https://github.com/gavv/httpexpect/issues/274

## [v2.14.0][v2.14.0] - 04 Mar 2023

* Bump minimal Go version to 1.16
* Bump golang.org/x/net to 0.7.0
* Fix handling of DefaultFormatter.LineWidth
* Update documentation
* Improve test coverage
* Refactor and cleanup tests

[v2.14.0]: https://github.com/gavv/httpexpect/releases/tag/v2.14.0

## [v2.13.0][v2.13.0] - 23 Feb 2023

* Rename `Object.ValueEqual` to `HasValue` ([gh-252][gh-252])
* Add `Array.HasValue` and `NotHasValue` ([gh-286][gh-286])
* Rename `Array.Element` to `Value` ([gh-252][gh-252])
* Add `Number.IsInt`, `IsUint`, `IsFinite` ([gh-155][gh-155])
* Add `Environment.List` and `Glob` ([gh-259][gh-259])
* Add `Environment.Clear` method ([gh-260][gh-260])
* Deprecate `Array.First` and `Last`
* Preparations to make lib thread-safe ([gh-233][gh-233])
* Improve documentation
* Improve test coverage
* Refactor and cleanup tests

[v2.13.0]: https://github.com/gavv/httpexpect/releases/tag/v2.13.0

[gh-155]: https://github.com/gavv/httpexpect/issues/155
[gh-233]: https://github.com/gavv/httpexpect/issues/233
[gh-252]: https://github.com/gavv/httpexpect/issues/252
[gh-259]: https://github.com/gavv/httpexpect/issues/259
[gh-260]: https://github.com/gavv/httpexpect/issues/260
[gh-286]: https://github.com/gavv/httpexpect/issues/286

## [v2.12.0][v2.12.0] - 08 Feb 2023

* Rename `Value.Null` to `IsNull` ([gh-252][gh-252])
* Rename `Boolean.True` and `Boolean.False` to `IsTrue` and `IsFalse` ([gh-252][gh-252])
* Correctly fill `AssertionContext.Response` field
* Panic on invalid `AssertionHandler`
* Add `String.InListFold` and `NotInListFold` ([gh-261][gh-261])
* Deprecate `RetryTemporaryNetworkErrors` and `RetryTemporaryNetworkAndServerErrors`, add `RetryTimeoutErrors` and `RetryTimeoutAndServerErrors` ([gh-270][gh-270])
* Lazy reading of response body ([gh-244][gh-244])
* Preparations to make lib thread-safe ([gh-233][gh-233])
* Improve documentation
* Improve test coverage

[v2.12.0]: https://github.com/gavv/httpexpect/releases/tag/v2.12.0

[gh-233]: https://github.com/gavv/httpexpect/issues/233
[gh-244]: https://github.com/gavv/httpexpect/issues/244
[gh-252]: https://github.com/gavv/httpexpect/issues/252
[gh-261]: https://github.com/gavv/httpexpect/issues/261
[gh-270]: https://github.com/gavv/httpexpect/issues/270

## [v2.11.0][v2.11.0] - 02 Feb 2023

* Add `FatalReporter` struct ([gh-187][gh-187])
* Add `Environment.Delete` method ([gh-235][gh-235])
* Add `Value.Decode` method ([gh-192][gh-192])
* Rename `Empty` to `IsEmpty` ([gh-252][gh-252])
* Rename `Equal` to `IsEqual` ([gh-252][gh-252])
* Add `DefaultFormatter.FloatFormat` option, rework formatting of numbers ([gh-190][gh-190])
* Add `InList` method ([gh-250][gh-250])
* Add `Value.IsXxx` methods ([gh-253][gh-253])
* Improve documentation
* Improve test coverage

[v2.11.0]: https://github.com/gavv/httpexpect/releases/tag/v2.11.0

[gh-187]: https://github.com/gavv/httpexpect/issues/187
[gh-190]: https://github.com/gavv/httpexpect/issues/190
[gh-192]: https://github.com/gavv/httpexpect/issues/192
[gh-235]: https://github.com/gavv/httpexpect/issues/235
[gh-250]: https://github.com/gavv/httpexpect/issues/250
[gh-252]: https://github.com/gavv/httpexpect/issues/252
[gh-253]: https://github.com/gavv/httpexpect/issues/253

## [v2.10.0][v2.10.0] - 28 Jan 2023

* Add `Decode` method ([gh-192][gh-192])
* Add `Alias` method ([gh-171][gh-171])
* Rename `Array.Contains` to `Array.ContainsAll` ([gh-252][gh-252])
* Rename `Array.Elements` to `Array.ConsistsOf` ([gh-252][gh-252])
* Rename `Cookie.HaveMaxAge` to `Cookie.HasMaxAge` ([gh-252][gh-252])
* Rename `DateTime.GetXxx` to `DateTime.Xxx` ([gh-252][gh-252])
* Preparations to make lib thread-safe ([gh-229][gh-229])
* Check call order of `Request` methods ([gh-162][gh-162])
* Improve documentation
* Improve test coverage

[v2.10.0]: https://github.com/gavv/httpexpect/releases/tag/v2.10.0

[gh-162]: https://github.com/gavv/httpexpect/issues/162
[gh-171]: https://github.com/gavv/httpexpect/issues/171
[gh-191]: https://github.com/gavv/httpexpect/issues/192
[gh-229]: https://github.com/gavv/httpexpect/issues/229
[gh-252]: https://github.com/gavv/httpexpect/issues/252
