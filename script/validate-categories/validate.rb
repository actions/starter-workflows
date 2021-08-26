require 'linguist'
require 'json'
require 'scout/tech_stack'

@languages = Linguist::Language.all.map(&:name)
@tech_stacks = Scout::TechStack.all.map(&:name)

settings = JSON.parse(File.read('./settings.json'))
folders = settings['folders']
@allowed_categories = settings['allowed_categories']

def validateCategories(categories)
    categoryErrors = []
    categories && categories.each do |category|
        if ! @allowed_categories.include?(category) && !@languages.include?(category) && !@tech_stacks.include?(category)
            categoryErrors.push(category)
        end
    end
    return categoryErrors
end

result = []
for folder in folders
    files = Dir.entries(folder).select {|entry| File.file?(File.join(folder, entry)) && (File.extname(entry) == ".yaml" || File.extname(entry) == ".yml") }
    for file in files
        workflowId = File.basename(file,  File.extname(file))
        errors = []
        propertiesPath = folder + "/" + "properties/" + workflowId+".properties.json"
        if(File.exist?(propertiesPath)) 
            properties = JSON.parse(File.read(propertiesPath))
            categoriesErrors = validateCategories(properties['categories'])
            categoriesErrors.each { |e| errors.push(e) }
        else
            errors.push("properties file not found")
        end
        if errors.length > 0
            result.push({"id" => workflowId, "errors" => errors})
        end
    end
end
if result.length > 0
    generic_message = ":heavy_exclamation_mark: There are unrecognised categories found in the repo. Please note that using the right categories will result in better visibility and recommendation."
    header = "|Workflow Template Id|Unrecognised Categories|\n|:----:  |:----:  |"
    rows = result.map { |r| "|#{r["id"]}|#{r["errors"]}|"}
    puts "::set-output name=unrecognised-categories:: #{generic_message}\n#{header}\n#{rows.join("\n")}"
end
