#!/usr/bin/env ruby
# Runner-only UI test target; distribution project and scheme stay unchanged in Git.
require 'xcodeproj'

root = File.expand_path('..', __dir__)
project_path = File.join(root, 'apps/mobile/ios/StuffStash.xcodeproj')
project = Xcodeproj::Project.open(project_path)
app = project.targets.find { |target| target.name == 'StuffStash' }
abort 'Missing StuffStash application target' unless app
abort 'Audit target already exists' if project.targets.any? { |target| target.name == 'StuffStashAuditTests' }

target = project.new_target(:ui_test_bundle, 'StuffStashAuditTests', :ios, '15.1')
target.add_dependency(app)
source = project.main_group.new_file('../native-audit/OnboardingAuditTests.swift')
target.source_build_phase.add_file_reference(source)
target.build_configurations.each do |configuration|
  configuration.build_settings.merge!({
    'PRODUCT_BUNDLE_IDENTIFIER' => 'org.stuffstash.mobile.audit-tests',
    'SWIFT_VERSION' => '5.0',
    'GENERATE_INFOPLIST_FILE' => 'YES',
    'TEST_TARGET_NAME' => 'StuffStash',
    'TARGETED_DEVICE_FAMILY' => '1,2',
    'CODE_SIGNING_ALLOWED' => 'NO'
  })
end
project.save

scheme_path = File.join(project_path, 'xcshareddata/xcschemes/StuffStash.xcscheme')
scheme = Xcodeproj::XCScheme.new(scheme_path)
scheme.test_action.testables.each { |testable| testable.xml_element.remove }
scheme.add_build_target(target)
scheme.add_test_target(target)
scheme.test_action.build_configuration = 'Release'
scheme.save_as(project_path, 'StuffStashAudit', true)
