#!/bin/bash
# update-changelog.sh - Автоматическое обновление changelog

VERSION=$(cat VERSION)
PACKAGE="system-monitor"
DATE=$(date -R)
NAME=$(git config user.name 2>/dev/null || echo "Andrey Vorobev")
EMAIL=$(git config user.email 2>/dev/null || echo "andrey.vorobev.aqa@gmail.com")

# Создаем временный файл
TEMP_FILE=$(mktemp)

# Записываем новую запись
cat > $TEMP_FILE << EOF
$PACKAGE ($VERSION) unstable; urgency=medium

  * Version $VERSION release

 -- $NAME <$EMAIL>  $DATE

EOF

# Добавляем старый changelog если существует
if [ -f debian/changelog ]; then
    cat debian/changelog >> $TEMP_FILE
fi

# Заменяем changelog
mv $TEMP_FILE debian/changelog

echo "Changelog updated to version $VERSION"