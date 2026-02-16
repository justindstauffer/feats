.PHONY: ios-build ios-test ios-destinations

IOS_PROJECT := ios/Feats/Feats.xcodeproj
IOS_SCHEME := Feats
IOS_DESTINATION := platform=iOS Simulator,name=iPhone 17,OS=26.0.1
IOS_DERIVED_DATA := /tmp/feats-derived

ios-build:
	xcodebuild build-for-testing \
		-project $(IOS_PROJECT) \
		-scheme $(IOS_SCHEME) \
		-destination '$(IOS_DESTINATION)' \
		-derivedDataPath $(IOS_DERIVED_DATA) \
		CODE_SIGNING_ALLOWED=NO

ios-test:
	xcodebuild test \
		-project $(IOS_PROJECT) \
		-scheme $(IOS_SCHEME) \
		-destination '$(IOS_DESTINATION)' \
		-derivedDataPath $(IOS_DERIVED_DATA) \
		CODE_SIGNING_ALLOWED=NO \
		-only-testing:FeatsTests

ios-destinations:
	xcodebuild -showdestinations \
		-project $(IOS_PROJECT) \
		-scheme $(IOS_SCHEME)
