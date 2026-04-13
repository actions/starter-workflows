export function normalizeSvgIconName(iconName?: string): string | undefined {
  if (!iconName || iconName.startsWith("octicon")) {
    return iconName;
  }

  if (iconName.startsWith("lucide ")) {
    const lucideName = iconName.slice("lucide ".length).split(".")[0].trim();
    return lucideName ? `lucide-${lucideName}` : undefined;
  }

  return iconName;
}
