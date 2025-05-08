<p>
   English&nbsp ｜&nbsp <a href="allow_nil_CN.md">中文</a>
</p>

# allowNil Option: Gracefully Handling Optional Dependencies

- [allowNil Option: Gracefully Handling Optional Dependencies](#allownil-option-gracefully-handling-optional-dependencies)
  - [What is allowNil?](#what-is-allownil)
  - [Use Cases](#use-cases)
  - [How to Use](#how-to-use)
  - [Implementation Principle](#implementation-principle)
  - [Usage Examples](#usage-examples)
    - [Basic Usage](#basic-usage)
    - [With Conditional Logic](#with-conditional-logic)
  - [Best Practices](#best-practices)
  - [Considerations](#considerations)
  - [Summary](#summary)


When using the Gone dependency injection framework, sometimes we need to handle "optional" dependencies — dependencies that, if not properly injected, should allow the program to continue running rather than throwing an error. The Gone framework provides the `allowNil` option to meet this need.

## What is allowNil?

`allowNil` is a tag option in the Gone framework used to mark dependencies that "can be nil". When a field is marked with `allowNil`, if the framework cannot find or inject the corresponding dependency, the program will continue to execute rather than failing due to the missing dependency.

## Use Cases

The following scenarios are suitable for using the `allowNil` option:

1. **Optional Features**: Some features are optional, and the system can continue to work without them
2. **Environment Differences**: Different environments may have different dependency availability
3. **Progressive Migration**: During system refactoring, temporarily allowing some new dependencies to be nil
4. **Conditional Dependencies**: Based on configuration or runtime conditions, some dependencies may not need to be injected

## How to Use

In struct field tags, use `option:"allowNil"` to mark a field that can accept injection failure:
```go
type MyService struct {
    gone.Flag
    // Required dependency - injection failure will cause an error
    Required Logger gone:"logger"
    // Optional dependency - injection failure will not cause an error
    Optional Analytics gone:"analytics" option:"allowNil"
}
```

## Implementation Principle

When the Gone framework attempts to inject dependencies, it checks whether the field is marked with `option:"allowNil"`. If it is marked and the dependency injection fails (e.g., dependency not found, type mismatch, etc.), Gone will ignore this error and continue processing subsequent fields, rather than immediately returning an error and terminating the entire injection process.

At the code level, the implementation logic is as follows:

1. Parse the `option` value in the field tag
2. If the value is `allowNil`, set the `isAllowNil` flag
3. Check this flag at various possible failure points in the injection process
4. If `isAllowNil` is true and injection fails, skip the error and continue execution

## Usage Examples

### Basic Usage
```go
type Application struct {
    gone.Flag
    // Core service - required
    DB Database gone:"database"
    // Optional monitoring service
    Monitoring MonitorService gone:"monitor" option:"allowNil"
    // Optional cache service
    Cache CacheService gone:"cache" option:"allowNil"
}
func (a Application) Init() error {
    // Check if optional dependencies were successfully injected
    if a.Monitoring != nil {
    // Use monitoring service
    }
    if a.Cache != nil {
    // Use cache service
    }
    return nil
}
```

### With Conditional Logic
```go
type API struct {
    gone.Flag
    // Optional rate limiter
    RateLimiter Limiter gone:"rateLimiter" option:"allowNil"
}
func (a API) HandleRequest(request Request) Response {
    // If rate limiter is available, use it
    if a.RateLimiter != nil {
        if !a.RateLimiter.Allow(request) {
            return Response{Status: 429, Message: "Too many requests"}
        }
    }
    // Process request...
    return Response{Status: 200, Message: "Success"}
}
```

## Best Practices

1. **Use Cautiously**: Only use `allowNil` for truly optional dependencies
2. **Always Check**: Always check if a dependency marked as `allowNil` is nil before using it
3. **Provide Default Behavior**: For optional dependencies, provide a reasonable default behavior
4. **Clear Documentation**: In code comments, explain which dependencies are optional and their purpose

## Considerations

- `allowNil` only applies to the dependency injection process and does not affect other aspects of the struct
- For basic types (such as int, string), even if marked with `allowNil`, they will not be nil but will maintain their zero values
- In concurrent environments, be aware that optional dependencies may become nil at any time

## Summary

The `allowNil` option is a powerful feature provided by the Gone framework that allows us to gracefully handle optional dependencies, improving the robustness and flexibility of the system. By using this feature appropriately, we can build applications that are more adaptive and fault-tolerant.