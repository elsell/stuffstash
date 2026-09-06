# Exercise the real Xcode project adapter on a disposable project, without a build.
require 'xcodeproj'
require 'tmpdir'
require 'fileutils'

root = File.expand_path('..', __dir__)
Dir.mktmpdir do |directory|
  relative = 'apps/mobile/ios/StuffStash.xcodeproj'
  FileUtils.mkdir_p(File.join(directory, 'apps/mobile/ios'))
  FileUtils.cp_r(File.join(root, relative), File.join(directory, relative))
  before = Xcodeproj::Project.open(File.join(directory, relative))
  environment = { 'STUFF_STASH_SIGNING_IDENTITY' => 'A' * 40,
                  'STUFF_STASH_SIGNING_PROFILE' => '12345678-1234-1234-1234-123456789abc' }
  success = system(environment, RbConfig.ruby, File.join(root, 'scripts/configure-ios-signing.rb'), chdir: directory)
  raise 'Signing configuration failed' unless success
  after = Xcodeproj::Project.open(File.join(directory, relative))
  before.build_configurations.each do |configuration|
    raise 'Changed project signing defaults' unless configuration.build_settings == after.objects_by_uuid.fetch(configuration.uuid).build_settings
  end
  before.targets.each do |target|
    target.build_configurations.each do |configuration|
      expected = configuration.build_settings.dup
      if target.name == 'StuffStash' && configuration.name == 'Release'
        expected.merge!('CODE_SIGN_STYLE' => 'Manual',
                        'CODE_SIGN_IDENTITY' => environment.fetch('STUFF_STASH_SIGNING_IDENTITY'),
                        'PROVISIONING_PROFILE_SPECIFIER' => environment.fetch('STUFF_STASH_SIGNING_PROFILE'))
      end
      raise "Unexpected settings change: #{target.name}/#{configuration.name}" unless expected == after.objects_by_uuid.fetch(configuration.uuid).build_settings
    end
  end
end
