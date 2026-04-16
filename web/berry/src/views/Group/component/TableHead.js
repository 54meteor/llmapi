import PropTypes from 'prop-types';
import { TableHead, TableRow, TableCell, TableSortLabel, Box } from '@mui/material';
import { moduleName } from 'utils/common';

const headCells = [
  {
    id: 'id',
    numeric: true,
    label: 'ID',
    sort: true
  },
  {
    id: 'name',
    numeric: false,
    label: '分组名称',
    sort: true
  },
  {
    id: 'ratio',
    numeric: true,
    label: '费率倍率',
    sort: true
  },
  {
    id: 'created_at',
    numeric: true,
    label: '创建时间',
    sort: true
  },
  {
    id: 'updated_at',
    numeric: true,
    label: '更新时间',
    sort: true
  },
  {
    id: 'action',
    numeric: false,
    label: '操作',
    sort: false
  }
];

export default function GroupTableHead({ order, orderBy, onRequestSort }) {
  const createSortHandler = (property) => (event) => {
    onRequestSort(event, property);
  };

  return (
    <TableHead>
      <TableRow>
        {headCells.map((headCell) => (
          <TableCell
            key={headCell.id}
            align={headCell.numeric ? 'right' : 'left'}
            sortDirection={orderBy === headCell.id ? order : false}
            sx={{ fontWeight: 600 }}
          >
            {headCell.sort ? (
              <TableSortLabel
                active={orderBy === headCell.id}
                direction={orderBy === headCell.id ? order : 'asc'}
                onClick={createSortHandler(headCell.id)}
              >
                {headCell.label}
              </TableSortLabel>
            ) : (
              headCell.label
            )}
          </TableCell>
        ))}
      </TableRow>
    </TableHead>
  );
}

GroupTableHead.propTypes = {
  order: PropTypes.string,
  orderBy: PropTypes.string,
  onRequestSort: PropTypes.func.isRequired
};