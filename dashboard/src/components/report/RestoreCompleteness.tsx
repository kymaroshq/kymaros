import { Card, CardHeader, CardTitle, CardBody } from '@/components/ui2/Card';
import { ResourceCard } from '@/components/ui2/ResourceCard';
import { Package, Network, Key, FileCode, HardDrive } from 'lucide-react';

interface Completeness {
  deployments?: string;
  services?: string;
  secrets?: string;
  configMaps?: string;
  pvcs?: string;
}

function parse(ratio?: string): [number, number] {
  if (!ratio) return [0, 0];
  const [present, expected] = ratio.split('/').map(Number);
  return [present || 0, expected || 0];
}

export function RestoreCompleteness({ completeness }: { completeness?: Completeness }) {
  if (!completeness) return null;

  const items = [
    { label: 'Deployments', icon: Package, ratio: completeness.deployments },
    { label: 'Services', icon: Network, ratio: completeness.services },
    { label: 'Secrets', icon: Key, ratio: completeness.secrets },
    { label: 'ConfigMaps', icon: FileCode, ratio: completeness.configMaps },
    { label: 'PVCs', icon: HardDrive, ratio: completeness.pvcs },
  ];

  return (
    <Card className="flex-1">
      <CardHeader>
        <CardTitle>Restore Completeness</CardTitle>
      </CardHeader>
      <CardBody>
        <div className="grid grid-cols-2 gap-2">
          {items.map((item) => {
            const [present, expected] = parse(item.ratio);
            return (
              <ResourceCard key={item.label} label={item.label} present={present} expected={expected} icon={item.icon} />
            );
          })}
        </div>
      </CardBody>
    </Card>
  );
}
