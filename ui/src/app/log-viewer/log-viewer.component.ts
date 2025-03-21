import {
  Component,
  Input,
  Pipe,
  PipeTransform,
  OnInit,
  OnChanges,
  SimpleChanges,
} from '@angular/core';
import { CommonModule } from '@angular/common';
import { LineChange, Log } from '../../types';
import { NzTableModule } from 'ng-zorro-antd/table';
import { NzInputModule } from 'ng-zorro-antd/input';
import { NzButtonModule } from 'ng-zorro-antd/button';
import { FormControl, ReactiveFormsModule } from '@angular/forms';
import { debounceTime, distinctUntilChanged } from 'rxjs/operators';
import { NzModalModule } from 'ng-zorro-antd/modal';

@Pipe({
  name: 'logLevel',
  pure: true,
  standalone: true,
})
export class LogLevelPipe implements PipeTransform {
  transform(value: string, idx: string, date: boolean = false): string {
    try {
      const info = JSON.parse(value);
      if (date) {
        return new Date(info[idx]).toLocaleString();
      }
      return info[idx]?.toString() ?? '';
    } catch (error) {
      return value;
    }
  }
}

interface LogFilter {
  search: string;
  level: string;
  startTime?: Date;
  endTime?: Date;
}

@Component({
  selector: 'app-log-viewer',
  standalone: true,
  imports: [
    CommonModule,
    LogLevelPipe,
    NzTableModule,
    NzInputModule,
    NzButtonModule,
    ReactiveFormsModule,
    NzModalModule,
  ],
  templateUrl: './log-viewer.component.html',
  styleUrl: './log-viewer.component.css',
})
export class LogViewerComponent implements OnInit, OnChanges {
  @Input() logs: LineChange[] = [];
  filteredLogs: LineChange[] = [];
  searchControl = new FormControl('');
  filter: LogFilter = {
    search: '',
    level: '',
  };
  dateRange: [Date, Date] | null = null;
  keys: string[] = [];

  constructor() {}

  ngOnInit() {
    this.filteredLogs = [...this.logs];

    // 设置搜索防抖
    this.searchControl.valueChanges
      .pipe(debounceTime(300), distinctUntilChanged())
      .subscribe((value) => {
        this.filterLogs(value || '');
      });

    this.filterLogs('');
  }

  // 监听输入属性变化
  ngOnChanges(changes: SimpleChanges) {
    if (changes['logs']) {
      // 当 logs 发生变化时
      this.filteredLogs = [...this.logs];
      // 重新应用当前的搜索过滤
      this.filterLogs(this.searchControl.value || '');
    }
  }

  private filterLogs(searchText: string) {
    if (searchText.trim() === '') {
      this.filteredLogs = [...this.logs];
    } else {
      this.filteredLogs = this.logs.filter((log) =>
        log.new_text.toLowerCase().includes(searchText.toLowerCase())
      );
    }

    let couner = new Map<string, number>();
    this.filteredLogs.forEach((element) => {
      try {
        const item = JSON.parse(element.new_text);
        for (const key in item) {
          if (!couner.has(key)) {
            // this.keys.push(key);
            couner.set(key, 0);
          } else {
            couner.set(key, couner.get(key)! + 1);
          }
        }
      } catch (error) {
        console.log(element.new_text);
      }
    });
    console.log(couner);
    this.keys = [];
    for (const [key, value] of couner.entries()) {
      if (value > 5) {
        this.keys.push(key);
      }
    }
    console.log(this.keys);
  }

  onRowClick(log: LineChange) {
    console.log(log);
  }
}
