Pod::Spec.new do |s|
  s.name = 'StuffStashArchiveTransfer'
  s.version = '1.0.0'
  s.summary = 'Stream portable archive uploads from local files'
  s.description = 'Bounded-memory archive upload with cancellation and redirect rejection.'
  s.homepage = 'https://github.com/elsell/stuffstash'
  s.license = { :type => 'UNLICENSED' }
  s.author = 'Stuff Stash contributors'
  s.source = { :git => 'https://github.com/elsell/stuffstash.git' }
  s.platforms = { :ios => '15.1' }
  s.swift_version = '5.9'
  s.static_framework = true
  s.dependency 'ExpoModulesCore'
  s.source_files = '**/*.swift'
end
