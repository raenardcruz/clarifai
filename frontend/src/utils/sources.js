export const SOURCE_META = {
  local: {
    key: 'local',
    label: 'ClarifAi',
    shortLabel: 'ClarifAi',
    isImported: false,
    color: '#0071E3',
    bgColor: 'rgba(0, 113, 227, 0.08)',
    borderColor: 'rgba(0, 113, 227, 0.2)'
  },
  fathom: {
    key: 'fathom',
    label: 'Fathom',
    shortLabel: 'Fathom',
    isImported: true,
    logoUrl: '/logos/fathom.svg',
    iconUrl: '/logos/fathom-icon.svg',
    color: '#0891B2',
    bgColor: 'rgba(8, 145, 178, 0.1)',
    borderColor: 'rgba(8, 145, 178, 0.28)'
  },
  fireflies: {
    key: 'fireflies',
    label: 'Fireflies.ai',
    shortLabel: 'Fireflies',
    isImported: true,
    logoUrl: '/logos/fireflies.svg',
    iconUrl: '/logos/fireflies-icon.svg',
    color: '#9333EA',
    bgColor: 'rgba(147, 51, 234, 0.1)',
    borderColor: 'rgba(147, 51, 234, 0.28)'
  },
  otter: {
    key: 'otter',
    label: 'Otter.ai',
    shortLabel: 'Otter',
    isImported: true,
    logoUrl: '/logos/otter.svg',
    iconUrl: '/logos/otter.svg',
    color: '#2563EB',
    bgColor: 'rgba(37, 99, 235, 0.1)',
    borderColor: 'rgba(37, 99, 235, 0.28)'
  }
}

export function getSourceMeta(source) {
  const key = (source || 'local').toLowerCase()
  return SOURCE_META[key] || SOURCE_META.local
}
