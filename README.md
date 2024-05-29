# Tolog
Just a log package. Have to create a logs dir in your project. 

## Usage
### Basic
```
    tolog.Debug("debug").PrintAndWriteSafe()
    tolog.Infof("info").PrintAndWriteSafe()
```

### Options
```
    tolog.Log(WithType("info"), WithContext("Info message")).PrintAndWriteSafe()
```

### Multiple
```
    tolog.Info("Info message").PrintAndWriteSafe()
    tolog.Infof("Infof message %s","string").PrintAndWriteSafe()
    tolog.Infoln("Infoln message", "this is message").PrintAndWriteSafe()
```

## Log level
- Info
- Warning
- Error
- Debug
- Notice
- Unknown

## Log setting
- logFileDateFormat
- logTimeFormat
- LogfilePrefix
- LogWithColor
- channelSize
- logTicker

## Log setting function
```
    SetLogWithColor()
    SetLogPrefix()
    SetLogChannelSize()
    SetLogTickerTime()
    SetLogFileDateFormat()
    SetLogTimeFormat()
```

## Print & Write
```
    PrintAndWriteSafe()
    WriteSafe()
    Print()
```