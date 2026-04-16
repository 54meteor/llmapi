import PropTypes from 'prop-types';
import { TableRow, TableCell, IconButton, Tooltip } from '@mui/material';
import { formatTimestamp } from 'utils/common';
import EditIcon from '@mui/icons-material/Edit';
import DeleteIcon from '@mui/icons-material/Delete';

const GroupTableRow = ({ row, onEdit, onDelete }) => {
  return (
    <TableRow hover>
      <TableCell align="right">{row.id}</TableCell>
      <TableCell>{row.name}</TableCell>
      <TableCell align="right">{row.ratio}</TableCell>
      <TableCell align="right">{formatTimestamp(row.created_at)}</TableCell>
      <TableCell align="right">{formatTimestamp(row.updated_at)}</TableCell>
      <TableCell>
        <Tooltip title="编辑">
          <IconButton onClick={() => onEdit(row)} size="small">
            <EditIcon />
          </IconButton>
        </Tooltip>
        <Tooltip title="删除">
          <IconButton onClick={() => onDelete(row)} size="small" color="error">
            <DeleteIcon />
          </IconButton>
        </Tooltip>
      </TableCell>
    </TableRow>
  );
};

export default GroupTableRow;

GroupTableRow.propTypes = {
  row: PropTypes.object.isRequired,
  onEdit: PropTypes.func.isRequired,
  onDelete: PropTypes.func.isRequired
};