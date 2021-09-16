require 'json'

settings = JSON.parse(File.read('./settings.json'))
folders = settings['folders']
@allowed_categories = settings['allowed_categories']

def validateCategories(categories)
    return categories.nil? || (categories.is_a?(Array) && ! categories.select{|l| @allowed_categories.detect{|p| p.casecmp(l) == 0 } }.empty? )
end

result = []
for folder in folders
    files = Dir.entries(folder).select {|entry| File.file?(File.join(folder, entry)) && (File.extname(entry) == ".yaml" || File.extname(entry) == ".yml") }
    for file in files
        workflowId = File.basename(file,  File.extname(file))
        raiseError = false
        propertiesPath = folder + "/" + "properties/" + workflowId+".properties.json"
        if(File.exist?(propertiesPath)) 
            begin
                properties = JSON.parse(File.read(propertiesPath))
                raiseError = ! validateCategories(properties['categories'])
            rescue JSON::ParserError
                raiseError = true
            end
        else
            raiseError = true 
        end
        if raiseError
            result.push({"id" => workflowId})
        end
    end
end
if result.length > 0
    result.each do |r|
        puts "::set-output name=unrecognised-categories-#{r["id"]}:: \|#{r["id"]}\|"
    end
end