# Changelog

## [v2.17.0][v2.17.0] - 05 Mar 2025

* Rename `Match.Index` to `Submatch` (#252)
* Rename `Match.Name` to `NamedSubmatch` (#252)
* Rename `Match.Values`/`NotValues` to `HasSubmatches`/`NotHasSubmatches` (#252)
* Rename `Cookie.HasMaxAge`/`NotContainsMaxAge` to `ContainsMaxAge`/`NotContainsMaxAge` (#252)
* Add `Response.Reader` for accessing response body directly (#382)
* Add `Request.WithRetryPolicyFunc` (#435)
* Add `Request.WithQueryEncoder` and `QueryEncoderFormKeepZeros` (#438)
* Add `Number.InDeltaRelative` (#306)
* Add `Object.Length`
* Update `Request.WithPath` to always format floats in decimal notation (#449)
* Fix panic in `DebugPrinter` when printing body (#444)
* Fix bugs in `NewMatchC()` and `NewWebsocketMessageC()`
* Improve test coverage
* Improve documentation and examples
* Add TLS example (#205)
* Bump some dependencies

[v2.17.0]: https://github.com/gavv/httpexpect/releases/tag/v2.17.0

## [v2.16.0][v2.16.0] - 03 Oct 2023

* Bump minimum Go version to 1.19
* Rename `Response.ContentType`/`ContentEncoding`/`TransferEncoding` to `HasXxx` (#252)
* Add `Request.WithReporter` and `WithAssertionHandler` (#234)
* Add stacktrace printing support (#160)
* Colorize JSON values (#334)
* Colorize HTTP requests and responses (#343)
* Prevent panic if `flag.Parse` hasn't been called (#410)
* Refactor and cleanup tests
* Improve test coverage
* Improve documentation
* Improve CI

[v2.16.0]: https://github.com/gavv/httpexpect/releases/tag/v2.16.0

## [v2.15.0][v2.15.0] - 04 Apr 2023

* Bump minimal Go version to 1.17
* Bump golang.org/x/net to 0.7.0
* Support colored output (#161, #335)
* Support thousand separation when formatting numbers (#274)
* Print HTTP request and response in failure message (#159)
* Add `RequestFactoryFunc`, `ClientFunc`, `WebsocketDialerFunc`, `ReporterFunc`, `LoggerFunc` (#249)
* Add `PanicReporter` (#248)
* Fix error messages in `AssertReporter` and `RequireReporter` (1e30c2736042d3b7e88bfff1948b74af4e8f1306)
* Update documentation
* Improve test coverage
* Refactor and cleanup tests

[v2.15.0]: https://github.com/gavv/httpexpect/releases/tag/v2.15.0

## [v2.14.0][v2.14.0] - 04 Mar 2023

* Bump minimal Go version to 1.16
* Bump golang.org/x/net to 0.7.0
* Fix handling of DefaultFormatter.LineWidth
* Update documentation
* Improve test coverage
* Refactor and cleanup tests

[v2.14.0]: https://github.com/gavv/httpexpect/releases/tag/v2.14.0

## [v2.13.0][v2.13.0] - 23 Feb 2023

* Rename `Object.ValueEqual` to `HasValue` (#252)
* Add `Array.HasValue` and `NotHasValue` (#286)
* Rename `Array.Element` to `Value` (#252)
* Add `Number.IsInt`, `IsUint`, `IsFinite` (#155)
* Add `Environment.List` and `Glob` (#259)
* Add `Environment.Clear` method (#260)
* Deprecate `Array.First` and `Last`
* Preparations to make lib thread-safe (#233)
* Improve documentation
* Improve test coverage
* Refactor and cleanup tests

[v2.13.0]: https://github.com/gavv/httpexpect/releases/tag/v2.13.0

## [v2.12.0][v2.12.0] - 08 Feb 2023

* Rename `Value.Null` to `IsNull` (#252)
* Rename `Boolean.True` and `Boolean.False` to `IsTrue` and `IsFalse` (#252)
* Correctly fill `AssertionContext.Response` field
* Panic on invalid `AssertionHandler`
* Add `String.InListFold` and `NotInListFold` (#261)
* Deprecate `RetryTemporaryNetworkErrors` and `RetryTemporaryNetworkAndServerErrors`, add `RetryTimeoutErrors` and `RetryTimeoutAndServerErrors` (#270)
* Lazy reading of response body (#244)
* Preparations to make lib thread-safe (#233)
* Improve documentation
* Improve test coverage

[v2.12.0]: https://github.com/gavv/httpexpect/releases/tag/v2.12.0

## [v2.11.0][v2.11.0] - 02 Feb 2023

* Add `FatalReporter` struct (#187)
* Add `Environment.Delete` method (#235)
* Add `Value.Decode` method (#192)
* Rename `Empty` to `IsEmpty` (#252)
* Rename `Equal` to `IsEqual` (#252)
* Add `DefaultFormatter.FloatFormat` option, rework formatting of numbers (#190)
* Add `InList` method (#250)
* Add `Value.IsXxx` methods (#253)
* Improve documentation
* Improve test coverage

[v2.11.0]: https://github.com/gavv/httpexpect/releases/tag/v2.11.0

## [v2.10.0][v2.10.0] - 28 Jan 2023

* Add `Decode` method (#192)
* Add `Alias` method (#171)
* Rename `Array.Contains` to `Array.ContainsAll` (#252)
* Rename `Array.Elements` to `Array.ConsistsOf` (#252)
* Rename `Cookie.HaveMaxAge` to `Cookie.HasMaxAge` (#252)
* Rename `DateTime.GetXxx` to `DateTime.Xxx` (#252)
* Preparations to make lib thread-safe (#229)
* Check call order of `Request` methods (#162)
* Improve documentation
* Improve test coverage

[v2.10.0]: https://github.com/gavv/httpexpect/releases/tag/v2.10.0
