#!/bin/bash
# update-changelog.sh - Update the changelog automatically.

VERSION=$(cat VERSION)
PACKAGE="dockpulse"
DATE=$(date -R)
NAME=$(git config user.name 2>/dev/null || echo "Andrey Vorobev")
EMAIL=$(git config user.email 2>/dev/null || echo "andrey.vorobev.aqa@gmail.com")

# Create a temporary file.
TEMP_FILE=$(mktemp)

# Write the new entry.
cat > $TEMP_FILE << EOF
$PACKAGE ($VERSION) unstable; urgency=medium

  * Version $VERSION release

 -- $NAME <$EMAIL>  $DATE

EOF

# Append the existing changelog when present.
if [ -f debian/changelog ]; then
    cat debian/changelog >> $TEMP_FILE
fi

# Replace the changelog.
mv $TEMP_FILE debian/changelog

echo "Changelog updated to version $VERSION"
