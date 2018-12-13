require 'asciidoctor'

ignore! /.git/

guard 'shell' do
	watch(%r{^.+\.adoc$}) {|m|
		Asciidoctor.convert_file m[0]
	}
	watch(%r{doc/.+\.adoc$}) {|m|
		Asciidoctor.convert_file m[0]
	}
end

guard 'livereload' do
	watch(%r{^.+\.(css|js|html)$})
end
