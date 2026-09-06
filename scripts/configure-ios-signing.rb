# CocoaPods supplies the xcodeproj library used by the pinned CocoaPods toolchain. Only configure the app.
require 'xcodeproj'
project = Xcodeproj::Project.open('apps/mobile/ios/StuffStash.xcodeproj')
target = project.targets.find { |candidate| candidate.name == 'StuffStash' }
raise 'Missing StuffStash target' unless target
configuration = target.build_configurations.find { |candidate| candidate.name == 'Release' }
raise 'Missing Release configuration' unless configuration
configuration.build_settings['CODE_SIGN_STYLE'] = 'Manual'
configuration.build_settings['CODE_SIGN_IDENTITY'] = ENV.fetch('STUFF_STASH_SIGNING_IDENTITY')
configuration.build_settings['PROVISIONING_PROFILE_SPECIFIER'] = ENV.fetch('STUFF_STASH_SIGNING_PROFILE')
project.save
