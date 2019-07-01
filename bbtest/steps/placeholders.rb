
placeholder :path do
  match(/((?:\/[a-z0-9]+[a-z0-9(\/)(\-)]{1,100}[\w,\s-]+(\.?[A-Za-z0-9_-]{0,100})+))/) do |path|
    path
  end
end

placeholder :permissions do
  match(/-?[r-][w-][x-][r-][w-][x-][r-][w-][x-]/) do |permissions|
    perm = permissions.reverse[0...9].reverse.chars.each_slice(3).map { |part|
      (part[0] == 'r' ? 4 : 0) +
      (part[1] == 'w' ? 2 : 0) +
      (part[2] == 'x' ? 1 : 0)
    }.join('')

    "0#{perm}"
  end
end

