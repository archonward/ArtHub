type NoticeProps = {
  children: string;
  tone?: "error" | "info";
};

export default function Notice({ children, tone = "info" }: NoticeProps) {
  return <div className={`notice notice--${tone}`}>{children}</div>;
}
