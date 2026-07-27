import { PageContainer } from '@/components/shared'
import DataBackupSection from './DataBackupSection'
import BatchUpdateSection from './BatchUpdateSection'

export default function DataManagementPage() {
  return (
    <PageContainer>
      <div className="flex flex-col gap-6">
        <DataBackupSection />
        <BatchUpdateSection />
      </div>
    </PageContainer>
  )
}
