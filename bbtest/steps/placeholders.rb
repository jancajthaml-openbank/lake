
placeholder :path do
  match(/((?:\/[a-z0-9]+[a-z0-9(\/)(\-)]{1,100}[\w,\s-]+(\.?[A-Za-z0-9_-]{0,100})+))/) do |path|
    path
  end
end
