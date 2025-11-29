@echo off
set GRADLE_USER_HOME=%USERPROFILE%\.gradle
set JAVA_HOME=C:\Program Files\Java\jdk-17
set PATH=%JAVA_HOME%\bin;%PATH%
gradlew.bat %*
