Pod::Spec.new do |s|
  s.name = 'StuffStashColorWell'
  s.version = '1.0.0'
  s.summary = 'System color selection for Stuff Stash'
  s.description = 'A direct UIKit color well preserving optional RGB draft selection.'
  s.homepage = 'https://github.com/elsell/stuffstash'
  s.license = { :type => 'UNLICENSED' }
  s.author = 'Stuff Stash contributors'
  s.source = { :git => 'https://github.com/elsell/stuffstash.git' }
  s.platforms = { :ios => '15.1' }
  s.swift_version = '5.9'
  s.static_framework = true
  s.dependency 'ExpoModulesCore'
  s.source_files = '**/*.{h,m,mm,swift}'
end
