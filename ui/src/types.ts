export interface Log {
  path: string;
  modified: boolean;
  content: string;
  prev_content: string;
  line_changes: LineChange[];
}

export interface LineChange {
  type: string;
  old_line: number;
  new_line: number;
  old_text: string;
  new_text: string;
}
